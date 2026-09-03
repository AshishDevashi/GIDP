variable "name" {
  description = "Logical name of the DB instance; used to name AWS resources."
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type."
  type        = string
}

variable "ami_ssm_parameter" {
  description = "SSM public parameter that resolves to the AMI id to launch."
  type        = string
}

variable "key_name" {
  description = "Name of the EC2 key pair created for this instance."
  type        = string
}

variable "public_key" {
  description = "OpenSSH formatted public key registered with the instance."
  type        = string
}

variable "user_data" {
  description = "Cloud-init script rendered by the application and run on first boot."
  type        = string
}

variable "admin_secret_name" {
  description = "SSM parameter name holding the PostgreSQL admin password."
  type        = string
}

variable "admin_password" {
  description = "PostgreSQL admin password stored as an SSM SecureString."
  type        = string
  sensitive   = true
}

variable "root_volume_size_gb" {
  description = "Size of the root EBS volume in GiB."
  type        = number
  default     = 10
}

variable "data_volume_size_gb" {
  description = "Size of the persistent PostgreSQL data EBS volume in GiB."
  type        = number
  default     = 20
}

variable "data_device_name" {
  description = "Block device name used to attach the data volume."
  type        = string
  default     = "/dev/sdf"
}

variable "postgres_port" {
  description = "TCP port exposed for PostgreSQL."
  type        = number
  default     = 5432
}

variable "ssh_ingress_cidr" {
  description = "CIDR allowed to reach the instance over SSH."
  type        = string
  default     = "0.0.0.0/0"
}

variable "postgres_ingress_cidr" {
  description = "CIDR allowed to reach PostgreSQL."
  type        = string
  default     = "0.0.0.0/0"
}

variable "associate_public_ip" {
  description = "Whether to attach a public IP to the instance."
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags applied to every created resource."
  type        = map(string)
  default     = {}
}
