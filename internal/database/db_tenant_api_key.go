package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

const (
	DefaultPageLimit = 100
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (db *Db) CreateTenantApiKey(ctx context.Context, tenantId uint64, apiKey *lead_scraper_servicev1.TenantAPIKey) (*lead_scraper_servicev1.TenantAPIKey, error) {
	var (
		tQop = db.QueryOperator.TenantORM
	)
	
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 {
		return nil, ErrInvalidInput
	}

	if apiKey == nil {
		return nil, ErrInvalidInput
	}

	// ensure the tenant exists
	tenant, err := tQop.WithContext(ctx).Where(tQop.Id.Eq(tenantId)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}	

	apiKeyORM, err := apiKey.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM: %w", err)
	}

	// append api key to tenant
	if  err := tQop.ApiKeys.WithContext(ctx).Model(tenant).Append(&apiKeyORM); err != nil {
		return nil, err
	}

	// update the tenant model
	if _, err := tQop.WithContext(ctx).Where(tQop.Id.Eq(tenantId)).Updates(&tenant); err != nil {
		return nil, err
	}

	// convert to protobuf
	pbResult, err := apiKeyORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

func (db *Db) GetTenantApiKey(ctx context.Context, tenantId uint64, apiKeyId uint64) (*lead_scraper_servicev1.TenantAPIKey, error) {
	var (
		tQop = db.QueryOperator.TenantAPIKeyORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 || apiKeyId == 0 {
		return nil, ErrInvalidInput
	}
	
	res, err := tQop.WithContext(ctx).Where(tQop.TenantId.Eq(tenantId), tQop.Id.Eq(apiKeyId)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant api key: %w", err)
	}

	pbResult, err := res.ToPB(ctx)	
	if err != nil {
		return nil, err
	}
	
	return &pbResult, nil
}

func (db *Db) UpdateTenantApiKey(ctx context.Context, tenantId uint64, apiKey *lead_scraper_servicev1.TenantAPIKey) (*lead_scraper_servicev1.TenantAPIKey, error) {
	var (
		tQop = db.QueryOperator.TenantAPIKeyORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 || apiKey == nil || apiKey.Id == 0 {
		return nil, ErrInvalidInput
	}

	// Convert to ORM model
	apiKeyORM, err := apiKey.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM: %w", err)
	}

	// Update the API key
	result, err := tQop.WithContext(ctx).Where(
		tQop.TenantId.Eq(tenantId),
		tQop.Id.Eq(apiKey.Id),
	).Updates(&apiKeyORM)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant api key: %w", err)
	}

	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	// Get the updated record
	updated, err := tQop.WithContext(ctx).Where(
		tQop.TenantId.Eq(tenantId),
		tQop.Id.Eq(apiKey.Id),
	).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get updated tenant api key: %w", err)
	}

	// Convert back to protobuf
	pbResult, err := updated.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

func (db *Db) DeleteTenantApiKey(ctx context.Context, tenantId uint64, apiKeyId uint64, deletionType DeletionType) error {
	var (
		tQop = db.QueryOperator.TenantAPIKeyORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 || apiKeyId == 0 {
		return ErrInvalidInput
	}

	queryRef := tQop.WithContext(ctx)
	if deletionType == DeletionTypeSoft {
		queryRef = queryRef.Where(tQop.Id.Eq(apiKeyId), tQop.TenantId.Eq(tenantId))
	} else {
		queryRef = queryRef.Where(tQop.Id.Eq(apiKeyId), tQop.TenantId.Eq(tenantId)).Unscoped()
	}

	result, err := queryRef.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete tenant api key: %w", err)
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (db *Db) ListTenantApiKeys(ctx context.Context, tenantId uint64, limit int, offset int) ([]*lead_scraper_servicev1.TenantAPIKey, error) {
	var (
		tQop = db.QueryOperator.TenantAPIKeyORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if tenantId == 0 {
		return nil, ErrInvalidInput
	}

	// Apply pagination defaults if not specified
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if offset < 0 {
		offset = 0
	}

	// Get API keys for the tenant
	apiKeys, err := tQop.WithContext(ctx).
		Where(tQop.TenantId.Eq(tenantId)).
		Offset(offset).
		Limit(limit).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant api keys: %w", err)
	}

	// Convert to protobuf
	var result []*lead_scraper_servicev1.TenantAPIKey
	for _, apiKey := range apiKeys {
		pbApiKey, err := apiKey.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		result = append(result, &pbApiKey)
	}

	return result, nil
}