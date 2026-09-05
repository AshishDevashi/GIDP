package databases

// CreateRequest defines the payload for provisioning a new database.
type CreateRequest struct {
	DBInstanceID string `json:"db_instance_id,omitempty"`
	Name         string `json:"name" binding:"required,min=1,max=63"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password" binding:"required,min=6,max=128"`
	SizeMB       int32  `json:"size_mb" binding:"required,min=10,max=2000"`
}

// Response represents a persisted managed database.
type Response struct {
	ID               string `json:"id"`
	DBInstanceID     string `json:"db_instance_id"`
	Name             string `json:"name"`
	Username         string `json:"username"`
	AllocatedMB      int32  `json:"allocated_mb"`
	Status           string `json:"status"`
	ConnectionString string `json:"connection_string,omitempty"`
	CreatedBy        string `json:"created_by,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ConnectionStringResponse represents the connection parameters and DSN for a database.
type ConnectionStringResponse struct {
	ConnectionString string `json:"connection_string"`
	Host             string `json:"host"`
	Port             int32  `json:"port"`
	Database         string `json:"database"`
	Username         string `json:"username"`
	Password         string `json:"password,omitempty"`
}

// QuotaResponse represents the allocation state of the dbinstance storage pool.
type QuotaResponse struct {
	TotalCapacityMB int64 `json:"total_capacity_mb"`
	AllocatedMB     int64 `json:"allocated_mb"`
	AvailableMB     int64 `json:"available_mb"`
}
