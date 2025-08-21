package carddetection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"trusioo_api/internal/carddetection/entities"
	"trusioo_api/pkg/carddetection"
	"trusioo_api/pkg/database"
)

// Repository 卡片检测记录数据访问接口
type Repository interface {
	// 基础CRUD操作
	Create(record *entities.CardDetectionRecord) error
	Update(record *entities.CardDetectionRecord) error
	GetByID(id int64) (*entities.CardDetectionRecord, error)
	GetByRequestID(requestID string) (*entities.CardDetectionRecord, error)
	
	// 查询操作
	GetByUserID(userID int64, limit, offset int) ([]*entities.CardDetectionRecord, error)
	GetByUserIDAndStatus(userID int64, status string, limit, offset int) ([]*entities.CardDetectionRecord, error)
	GetByUserIDAndProductMark(userID int64, productMark carddetection.ProductMark, limit, offset int) ([]*entities.CardDetectionRecord, error)
	GetByCardNumber(userID int64, cardNumber string) ([]*entities.CardDetectionRecord, error)
	
	// CD产品和区域查询
	GetCDProducts() ([]*entities.CDProduct, error)
	GetCDRegions(productMark *string) ([]*entities.CDRegion, error)
	GetCDProductByMark(productMark string) (*entities.CDProduct, error)
	
	// 统计操作
	CountByUserID(userID int64) (int, error)
	CountByUserIDAndStatus(userID int64, status string) (int, error)
	GetSummaryByUserID(userID int64) (*entities.CardDetectionSummary, error)
	
	// 批量操作
	CreateBatch(records []*entities.CardDetectionRecord) error
	UpdateStatus(id int64, status string, checkResult interface{}, errorMessage string) error
}

// repository Repository接口的实现
type repository struct{}

// NewRepository 创建新的Repository实例
func NewRepository() Repository {
	return &repository{}
}

// Create 创建新的检测记录
func (r *repository) Create(record *entities.CardDetectionRecord) error {
	query := `
		INSERT INTO card_detection_records (
			user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`
	
	var checkResultJSON *string
	if record.CheckResult != nil {
		jsonData, err := json.Marshal(*record.CheckResult)
		if err != nil {
			return fmt.Errorf("failed to marshal check result: %w", err)
		}
		jsonStr := string(jsonData)
		checkResultJSON = &jsonStr
	}
	
	return database.DB.QueryRow(query,
		record.UserID,
		record.RequestID,
		record.CardNumber,
		record.PinCode,
		record.ProductMark,
		record.RegionID,
		record.RegionName,
		record.AutoType,
		record.CheckStatus,
		checkResultJSON,
		record.ErrorMessage,
		record.ResponseCode,
		record.ResponseTime,
		record.CheckedAt,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
}

// Update 更新检测记录
func (r *repository) Update(record *entities.CardDetectionRecord) error {
	query := `
		UPDATE card_detection_records 
		SET check_status = $2, check_result = $3, error_message = $4, response_code = $5,
			response_time = $6, checked_at = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	
	var checkResultJSON *string
	if record.CheckResult != nil {
		jsonData, err := json.Marshal(*record.CheckResult)
		if err != nil {
			return fmt.Errorf("failed to marshal check result: %w", err)
		}
		jsonStr := string(jsonData)
		checkResultJSON = &jsonStr
	}
	
	return database.DB.QueryRow(query,
		record.ID,
		record.CheckStatus,
		checkResultJSON,
		record.ErrorMessage,
		record.ResponseCode,
		record.ResponseTime,
		record.CheckedAt,
	).Scan(&record.UpdatedAt)
}

// GetByID 根据ID获取检测记录
func (r *repository) GetByID(id int64) (*entities.CardDetectionRecord, error) {
	var record entities.CardDetectionRecord
	var checkResultJSON sql.NullString
	
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records WHERE id = $1`
	
	err := database.DB.QueryRow(query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.RequestID,
		&record.CardNumber,
		&record.PinCode,
		&record.ProductMark,
		&record.RegionID,
		&record.RegionName,
		&record.AutoType,
		&record.CheckStatus,
		&checkResultJSON,
		&record.ErrorMessage,
		&record.ResponseCode,
		&record.ResponseTime,
		&record.CheckedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	
	// 解析JSON检测结果
	if checkResultJSON.Valid && checkResultJSON.String != "" {
		result := checkResultJSON.String
		record.CheckResult = &result
	}
	
	return &record, nil
}

// GetByRequestID 根据请求ID获取检测记录
func (r *repository) GetByRequestID(requestID string) (*entities.CardDetectionRecord, error) {
	var record entities.CardDetectionRecord
	var checkResultJSON sql.NullString
	
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records WHERE request_id = $1`
	
	err := database.DB.QueryRow(query, requestID).Scan(
		&record.ID,
		&record.UserID,
		&record.RequestID,
		&record.CardNumber,
		&record.PinCode,
		&record.ProductMark,
		&record.RegionID,
		&record.RegionName,
		&record.AutoType,
		&record.CheckStatus,
		&checkResultJSON,
		&record.ErrorMessage,
		&record.ResponseCode,
		&record.ResponseTime,
		&record.CheckedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	
	// 解析JSON检测结果
	if checkResultJSON.Valid && checkResultJSON.String != "" {
		result := checkResultJSON.String
		record.CheckResult = &result
	}
	
	return &record, nil
}

// GetByUserID 获取用户的检测记录
func (r *repository) GetByUserID(userID int64, limit, offset int) ([]*entities.CardDetectionRecord, error) {
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`
	
	return r.scanRecords(query, userID, limit, offset)
}

// GetByUserIDAndStatus 获取用户指定状态的检测记录
func (r *repository) GetByUserIDAndStatus(userID int64, status string, limit, offset int) ([]*entities.CardDetectionRecord, error) {
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records 
		WHERE user_id = $1 AND check_status = $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4`
	
	return r.scanRecords(query, userID, status, limit, offset)
}

// GetByUserIDAndProductMark 获取用户指定产品类型的检测记录
func (r *repository) GetByUserIDAndProductMark(userID int64, productMark carddetection.ProductMark, limit, offset int) ([]*entities.CardDetectionRecord, error) {
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records 
		WHERE user_id = $1 AND product_mark = $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4`
	
	return r.scanRecords(query, userID, productMark, limit, offset)
}

// GetByCardNumber 获取指定卡号的检测记录
func (r *repository) GetByCardNumber(userID int64, cardNumber string) ([]*entities.CardDetectionRecord, error) {
	query := `
		SELECT id, user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at, created_at, updated_at
		FROM card_detection_records 
		WHERE user_id = $1 AND card_number = $2
		ORDER BY created_at DESC`
	
	return r.scanRecords(query, userID, cardNumber)
}

// CountByUserID 统计用户的检测记录总数
func (r *repository) CountByUserID(userID int64) (int, error) {
	var count int
	err := database.DB.Get(&count, "SELECT COUNT(*) FROM card_detection_records WHERE user_id = $1", userID)
	return count, err
}

// CountByUserIDAndStatus 统计用户指定状态的检测记录数
func (r *repository) CountByUserIDAndStatus(userID int64, status string) (int, error) {
	var count int
	err := database.DB.Get(&count, 
		"SELECT COUNT(*) FROM card_detection_records WHERE user_id = $1 AND check_status = $2", 
		userID, status)
	return count, err
}

// GetSummaryByUserID 获取用户的检测记录汇总
func (r *repository) GetSummaryByUserID(userID int64) (*entities.CardDetectionSummary, error) {
	query := `
		SELECT 
			user_id,
			COUNT(*) as total_checks,
			SUM(CASE WHEN check_status = 'completed' THEN 1 ELSE 0 END) as success_checks,
			SUM(CASE WHEN check_status = 'failed' THEN 1 ELSE 0 END) as failed_checks,
			MAX(checked_at) as last_check_at
		FROM card_detection_records 
		WHERE user_id = $1
		GROUP BY user_id`
	
	var summary entities.CardDetectionSummary
	err := database.DB.Get(&summary, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有记录，返回空汇总
			return &entities.CardDetectionSummary{
				UserID:        userID,
				TotalChecks:   0,
				SuccessChecks: 0,
				FailedChecks:  0,
				LastCheckAt:   nil,
			}, nil
		}
		return nil, err
	}
	
	return &summary, nil
}

// CreateBatch 批量创建检测记录
func (r *repository) CreateBatch(records []*entities.CardDetectionRecord) error {
	if len(records) == 0 {
		return nil
	}
	
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO card_detection_records (
			user_id, request_id, card_number, pin_code, product_mark, region_id,
			region_name, auto_type, check_status, check_result, error_message,
			response_code, response_time, checked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`
	
	for _, record := range records {
		var checkResultJSON *string
		if record.CheckResult != nil {
			jsonData, err := json.Marshal(*record.CheckResult)
			if err != nil {
				return fmt.Errorf("failed to marshal check result: %w", err)
			}
			jsonStr := string(jsonData)
			checkResultJSON = &jsonStr
		}
		
		err := tx.QueryRow(query,
			record.UserID,
			record.RequestID,
			record.CardNumber,
			record.PinCode,
			record.ProductMark,
			record.RegionID,
			record.RegionName,
			record.AutoType,
			record.CheckStatus,
			checkResultJSON,
			record.ErrorMessage,
			record.ResponseCode,
			record.ResponseTime,
			record.CheckedAt,
		).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
		
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

// UpdateStatus 更新检测状态和结果
func (r *repository) UpdateStatus(id int64, status string, checkResult interface{}, errorMessage string) error {
	now := time.Now()
	var checkResultJSON *string
	
	if checkResult != nil {
		jsonData, err := json.Marshal(checkResult)
		if err != nil {
			return fmt.Errorf("failed to marshal check result: %w", err)
		}
		jsonStr := string(jsonData)
		checkResultJSON = &jsonStr
	}
	
	query := `
		UPDATE card_detection_records 
		SET check_status = $2, check_result = $3, error_message = $4, 
			checked_at = $5, updated_at = NOW()
		WHERE id = $1`
	
	_, err := database.DB.Exec(query, id, status, checkResultJSON, errorMessage, now)
	return err
}

// scanRecords 扫描记录的辅助方法
func (r *repository) scanRecords(query string, args ...interface{}) ([]*entities.CardDetectionRecord, error) {
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*entities.CardDetectionRecord
	for rows.Next() {
		var record entities.CardDetectionRecord
		var checkResultJSON sql.NullString
		
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.RequestID,
			&record.CardNumber,
			&record.PinCode,
			&record.ProductMark,
			&record.RegionID,
			&record.RegionName,
			&record.AutoType,
			&record.CheckStatus,
			&checkResultJSON,
			&record.ErrorMessage,
			&record.ResponseCode,
			&record.ResponseTime,
			&record.CheckedAt,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// 解析JSON检测结果
		if checkResultJSON.Valid && checkResultJSON.String != "" {
			result := checkResultJSON.String
			record.CheckResult = &result
		}
		
		records = append(records, &record)
	}
	
	return records, rows.Err()
}

// GetCDProducts 获取所有CD产品列表
func (r *repository) GetCDProducts() ([]*entities.CDProduct, error) {
	query := `
		SELECT id, product_mark, product_name, requires_region, requires_pin, 
		       card_format, card_length_min, card_length_max, pin_length, 
		       validation_pattern, supports_auto_type, status, created_at, updated_at
		FROM cd_products
		WHERE status = 'active'
		ORDER BY id ASC`
	
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cd_products: %w", err)
	}
	defer rows.Close()
	
	var products []*entities.CDProduct
	for rows.Next() {
		product := &entities.CDProduct{}
		scanErr := rows.Scan(
			&product.ID,
			&product.ProductMark,
			&product.ProductName,
			&product.RequiresRegion,
			&product.RequiresPin,
			&product.CardFormat,
			&product.CardLengthMin,
			&product.CardLengthMax,
			&product.PinLength,
			&product.ValidationPattern,
			&product.SupportsAutoType,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan cd_product: %w", scanErr)
		}
		products = append(products, product)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cd_products rows: %w", err)
	}
	
	return products, nil
}

// GetCDRegions 获取CD区域列表，可选择按产品标识过滤
func (r *repository) GetCDRegions(productMark *string) ([]*entities.CDRegion, error) {
	query := `
		SELECT id, product_mark, region_id, region_name, region_name_en, 
		       status, sort_order, created_at, updated_at
		FROM cd_regions`
	
	var args []interface{}
	if productMark != nil {
		query += ` WHERE product_mark = $1 AND status = 'active'`
		args = append(args, *productMark)
	} else {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY sort_order ASC, id ASC`
	
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cd_regions: %w", err)
	}
	defer rows.Close()
	
	var regions []*entities.CDRegion
	for rows.Next() {
		region := &entities.CDRegion{}
		scanErr := rows.Scan(
			&region.ID,
			&region.ProductMark,
			&region.RegionID,
			&region.RegionName,
			&region.RegionNameEn,
			&region.Status,
			&region.SortOrder,
			&region.CreatedAt,
			&region.UpdatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan cd_region: %w", scanErr)
		}
		regions = append(regions, region)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cd_regions rows: %w", err)
	}
	
	return regions, nil
}

// GetCDProductByMark 根据产品标识获取CD产品
func (r *repository) GetCDProductByMark(productMark string) (*entities.CDProduct, error) {
	query := `
		SELECT id, product_mark, product_name, requires_region, requires_pin, 
		       card_format, card_length_min, card_length_max, pin_length, 
		       validation_pattern, supports_auto_type, status, created_at, updated_at
		FROM cd_products
		WHERE product_mark = $1 AND status = 'active'`
	
	product := &entities.CDProduct{}
	err := database.DB.QueryRow(query, productMark).Scan(
		&product.ID,
		&product.ProductMark,
		&product.ProductName,
		&product.RequiresRegion,
		&product.RequiresPin,
		&product.CardFormat,
		&product.CardLengthMin,
		&product.CardLengthMax,
		&product.PinLength,
		&product.ValidationPattern,
		&product.SupportsAutoType,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cd_product by mark: %w", err)
	}
	
	return product, nil
}