package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

func (db *Db) CreateWebhookConfig(ctx context.Context, tenantId uint64, webhook *lead_scraper_servicev1.WebhookConfig) (*lead_scraper_servicev1.WebhookConfig, error) {
	var (
		tQop = db.QueryOperator.WorkspaceORM
	)
	
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 {
		return nil, ErrInvalidInput
	}

	if webhook == nil {
		return nil, ErrInvalidInput
	}

	// ensure the tenant exists
	tenant, err := tQop.WithContext(ctx).Where(tQop.Id.Eq(tenantId)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}	

	webhookORM, err := webhook.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM: %w", err)
	}

	// append webhook to tenant
	if err := tQop.Webhooks.WithContext(ctx).Model(tenant).Append(&webhookORM); err != nil {
		return nil, err
	}

	// update the tenant model
	if _, err := tQop.WithContext(ctx).Where(tQop.Id.Eq(tenantId)).Updates(&tenant); err != nil {
		return nil, err
	}

	// convert to protobuf
	pbResult, err := webhookORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

func (db *Db) GetWebhookConfig(ctx context.Context, workspaceId uint64, webhookId uint64) (*lead_scraper_servicev1.WebhookConfig, error) {
	var (
		wQop = db.QueryOperator.WebhookConfigORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if workspaceId == 0 || webhookId == 0 {
		return nil, ErrInvalidInput
	}
	
	res, err := wQop.WithContext(ctx).Where(wQop.WorkspaceId.Eq(workspaceId), wQop.Id.Eq(webhookId)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook config: %w", err)
	}

	pbResult, err := res.ToPB(ctx)	
	if err != nil {
		return nil, err
	}
	
	return &pbResult, nil
}

func (db *Db) UpdateWebhookConfig(ctx context.Context, workspaceId uint64, webhook *lead_scraper_servicev1.WebhookConfig) (*lead_scraper_servicev1.WebhookConfig, error) {
	var (
		wQop = db.QueryOperator.WebhookConfigORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if workspaceId == 0 || webhook == nil || webhook.Id == 0 {
		return nil, ErrInvalidInput
	}

	// Convert to ORM model
	webhookORM, err := webhook.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM: %w", err)
	}

	// Update the webhook
	result, err := wQop.WithContext(ctx).Where(
		wQop.WorkspaceId.Eq(workspaceId),
		wQop.Id.Eq(webhook.Id),
	).Updates(&webhookORM)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook config: %w", err)
	}

	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	// Get the updated record
	updated, err := wQop.WithContext(ctx).Where(
		wQop.WorkspaceId.Eq(workspaceId),
		wQop.Id.Eq(webhook.Id),
	).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get updated webhook config: %w", err)
	}

	// Convert back to protobuf
	pbResult, err := updated.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

func (db *Db) DeleteWebhookConfig(ctx context.Context, workspaceId uint64, webhookId uint64, deletionType DeletionType) error {
	var (
		wQop = db.QueryOperator.WebhookConfigORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if workspaceId == 0 || webhookId == 0 {
		return ErrInvalidInput
	}

	queryRef := wQop.WithContext(ctx)
	if deletionType == DeletionTypeSoft {
		queryRef = queryRef.Where(wQop.Id.Eq(webhookId), wQop.WorkspaceId.Eq(workspaceId))
	} else {
		queryRef = queryRef.Where(wQop.Id.Eq(webhookId), wQop.WorkspaceId.Eq(workspaceId)).Unscoped()
	}

	result, err := queryRef.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (db *Db) ListWebhookConfigs(ctx context.Context, workspaceId uint64, limit int, offset int) ([]*lead_scraper_servicev1.WebhookConfig, error) {
	var (
		wQop = db.QueryOperator.WebhookConfigORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if workspaceId == 0 {
		return nil, ErrInvalidInput
	}

	// Apply pagination defaults if not specified
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if offset < 0 {
		offset = 0
	}

	// Get webhooks for the tenant
	webhooks, err := wQop.WithContext(ctx).
		Where(wQop.WorkspaceId.Eq(workspaceId)).
		Offset(offset).
		Limit(limit).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook configs: %w", err)
	}

	// Convert to protobuf
	var result []*lead_scraper_servicev1.WebhookConfig
	for _, webhook := range webhooks {
		pbWebhook, err := webhook.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		result = append(result, &pbWebhook)
	}

	return result, nil
} 