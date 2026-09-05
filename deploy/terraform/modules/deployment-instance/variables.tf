variable "name" {
  description = "Logical name of the deployment instance; used to name AWS resources."
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

variable "root_volume_size_gb" {
  description = "Size of the root EBS volume in GiB."
  type        = number
  default     = 20
}

variable "ssh_ingress_cidr" {
  description = "CIDR allowed to reach the instance over SSH."
  type        = string
}

variable "kube_api_ingress_cidr" {
  description = "CIDR allowed to reach the k3s API server."
  type        = string
}

variable "http_ingress_cidr" {
  description = "CIDR allowed to reach HTTP workloads."
  type        = string
}

variable "https_ingress_cidr" {
  description = "CIDR allowed to reach HTTPS workloads."
  type        = string
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