output "instance_id" {
  description = "Id of the provisioned EC2 instance."
  value       = aws_instance.this.id
}

output "availability_zone" {
  description = "Availability zone the instance runs in."
  value       = aws_instance.this.availability_zone
}

output "public_ip" {
  description = "Public IPv4 address of the instance."
  value       = aws_instance.this.public_ip
}

output "private_ip" {
  description = "Private IPv4 address of the instance."
  value       = aws_instance.this.private_ip
}

output "security_group_id" {
  description = "Id of the security group attached to the instance."
  value       = aws_security_group.this.id
}

output "data_volume_id" {
  description = "Id of the persistent data EBS volume."
  value       = aws_ebs_volume.data.id
}

output "key_name" {
  description = "Name of the EC2 key pair attached to the instance."
  value       = aws_key_pair.this.key_name
}

output "admin_secret_name" {
  description = "SSM parameter name holding the PostgreSQL admin password."
  value       = aws_ssm_parameter.admin_password.name
}
