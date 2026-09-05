output "instance_id" {
  description = "Id of the provisioned EC2 instance."
  value       = aws_instance.this.id
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