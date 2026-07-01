package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"gitops-lite/apps/deploy-worker/internal/events"
	"gitops-lite/apps/deploy-worker/internal/executor"
	"gitops-lite/apps/deploy-worker/internal/health"
	"gitops-lite/pkg/model"
	"gitops-lite/pkg/repository"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queue        string
	deployRepo   *repository.DeploymentRepo
	logRepo      *repository.LogRepo
	jobRepo      *repository.JobRepo
	tfExecutor   *executor.TerraformExecutor
	checker      *health.Checker
	eventClient  *events.Client
	seq          atomic.Int64
	done         chan struct{}
}

func NewConsumer(
	rabbitmqURL, queue string,
	deployRepo *repository.DeploymentRepo,
	logRepo *repository.LogRepo,
	jobRepo *repository.JobRepo,
	tfExecutor *executor.TerraformExecutor,
	checker *health.Checker,
	eventClient *events.Client,
) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("queue declare: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("qos: %w", err)
	}

	return &Consumer{
		conn:        conn,
		channel:     ch,
		queue:       queue,
		deployRepo:  deployRepo,
		logRepo:     logRepo,
		jobRepo:     jobRepo,
		tfExecutor:  tfExecutor,
		checker:     checker,
		eventClient: eventClient,
		done:        make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue, "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Info().Msg("worker started, waiting for deploy jobs...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	<-c.done
	return nil
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var payload model.DeployJobPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Error().Err(err).Msg("invalid job payload")
		msg.Nack(false, false)
		return
	}

	log.Info().Str("deploy_id", payload.DeploymentID).Str("app", payload.AppName).Msg("processing deploy job")

	c.deployRepo.UpdateStatus(ctx, payload.DeploymentID, model.StatusInProgress, nil)
	c.eventClient.PublishUpdate(payload.DeploymentID, string(model.StatusInProgress))

	if err := c.appendLog(ctx, payload.DeploymentID, "received", model.LogInfo, "Job received by worker"); err != nil {
		log.Error().Err(err).Msg("failed to write log")
	}

	workDir := filepath.Join(os.TempDir(), "gitops", payload.DeploymentID)

	defer os.RemoveAll(workDir)

	if err := c.tfExecutor.Prepare(workDir, payload.AppName, payload.ImageTag, payload.EnvVars); err != nil {
		c.failDeployment(ctx, msg, payload.DeploymentID, "prepare failed: "+err.Error())
		return
	}

	if err := c.appendLog(ctx, payload.DeploymentID, "terraform_init", model.LogInfo, "Initializing Terraform..."); err != nil {
		log.Error().Err(err).Msg("failed to write log")
	}

	result := c.tfExecutor.RunInit(workDir)
	if !result.Success {
		c.appendLog(ctx, payload.DeploymentID, "terraform_init", model.LogError, result.Error)
		c.failDeployment(ctx, msg, payload.DeploymentID, result.Error)
		return
	}
	c.appendLog(ctx, payload.DeploymentID, "terraform_init", model.LogInfo, "Terraform initialized")

	if err := c.appendLog(ctx, payload.DeploymentID, "terraform_plan", model.LogInfo, "Planning Terraform..."); err != nil {
		log.Error().Err(err).Msg("failed to write log")
	}

	result = c.tfExecutor.RunPlan(workDir)
	if !result.Success {
		c.appendLog(ctx, payload.DeploymentID, "terraform_plan", model.LogError, result.Error)
		c.failDeployment(ctx, msg, payload.DeploymentID, result.Error)
		return
	}
	c.appendLog(ctx, payload.DeploymentID, "terraform_plan", model.LogInfo, "Terraform plan completed")

	if err := c.appendLog(ctx, payload.DeploymentID, "terraform_apply", model.LogInfo, "Applying Terraform..."); err != nil {
		log.Error().Err(err).Msg("failed to write log")
	}

	result = c.tfExecutor.RunApply(workDir)
	if !result.Success {
		c.appendLog(ctx, payload.DeploymentID, "terraform_apply", model.LogError, result.Error)
		c.failDeployment(ctx, msg, payload.DeploymentID, result.Error)
		return
	}
	c.appendLog(ctx, payload.DeploymentID, "terraform_apply", model.LogInfo, "Terraform apply completed")

	if err := c.appendLog(ctx, payload.DeploymentID, "health_check", model.LogInfo, "Running health check..."); err != nil {
		log.Error().Err(err).Msg("failed to write log")
	}

	ok, checkMsg := c.checker.Check("localhost", 8080, "/health")
	if !ok {
		c.appendLog(ctx, payload.DeploymentID, "health_check", model.LogError, checkMsg)
		c.failDeployment(ctx, msg, payload.DeploymentID, checkMsg)
		return
	}
	c.appendLog(ctx, payload.DeploymentID, "health_check", model.LogInfo, checkMsg)

	c.deployRepo.UpdateStatus(ctx, payload.DeploymentID, model.StatusSuccess, nil)
	c.eventClient.PublishUpdate(payload.DeploymentID, string(model.StatusSuccess))
	c.appendLog(ctx, payload.DeploymentID, "completed", model.LogInfo, "Deploy completed successfully")

	msg.Ack(false)

	log.Info().Str("deploy_id", payload.DeploymentID).Msg("deploy completed")
}

func (c *Consumer) failDeployment(ctx context.Context, msg amqp.Delivery, deployID, errMsg string) {
	c.deployRepo.UpdateStatus(ctx, deployID, model.StatusFailed, &errMsg)
	c.eventClient.PublishUpdate(deployID, string(model.StatusFailed))
	c.appendLog(ctx, deployID, "failed", model.LogError, errMsg)
	msg.Nack(false, false)
	log.Error().Str("deploy_id", deployID).Msg(errMsg)
}

func (c *Consumer) appendLog(ctx context.Context, deployID, step string, level model.LogLevel, message string) error {
	seq := c.seq.Add(1)
	c.eventClient.PublishLog(deployID, step, string(level), message)
	return c.logRepo.Append(ctx, deployID, step, level, message, seq)
}

func (c *Consumer) Stop() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.done)
}
