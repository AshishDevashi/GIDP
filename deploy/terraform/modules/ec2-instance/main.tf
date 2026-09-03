terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }
}

# Latest Amazon Linux 2023 AMI for the target region, resolved at plan time so
# no AMI id has to be hard-coded per region.
data "aws_ssm_parameter" "ami" {
  name = var.ami_ssm_parameter
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_subnet" "selected" {
  id = sort(data.aws_subnets.default.ids)[0]
}

resource "aws_key_pair" "this" {
  key_name   = var.key_name
  public_key = var.public_key
  tags       = var.tags
}

# PostgreSQL superuser password. Kept in SSM so it is never baked into user_data
# (which is readable from the instance metadata service by anything on the box).
resource "aws_ssm_parameter" "admin_password" {
  name        = var.admin_secret_name
  description = "PostgreSQL admin password for ${var.name}"
  type        = "SecureString"
  value       = var.admin_password
  tags        = var.tags
}

data "aws_kms_alias" "ssm" {
  name = "alias/aws/ssm"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "read_admin_password" {
  statement {
    actions   = ["ssm:GetParameter"]
    resources = [aws_ssm_parameter.admin_password.arn]
  }

  statement {
    actions   = ["kms:Decrypt"]
    resources = [data.aws_kms_alias.ssm.target_key_arn]
  }
}

resource "aws_iam_role" "this" {
  name               = "${var.name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
  tags               = var.tags
}

resource "aws_iam_role_policy" "this" {
  name   = "${var.name}-read-admin-password"
  role   = aws_iam_role.this.id
  policy = data.aws_iam_policy_document.read_admin_password.json
}

resource "aws_iam_instance_profile" "this" {
  name = "${var.name}-profile"
  role = aws_iam_role.this.name
  tags = var.tags
}

resource "aws_security_group" "this" {
  name        = "${var.name}-sg"
  description = "GIDP managed security group for ${var.name}"
  vpc_id      = data.aws_vpc.default.id
  tags        = merge(var.tags, { Name = "${var.name}-sg" })

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc_security_group_ingress_rule" "ssh" {
  security_group_id = aws_security_group.this.id
  description       = "SSH access"
  cidr_ipv4         = var.ssh_ingress_cidr
  from_port         = 22
  to_port           = 22
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "postgres" {
  security_group_id = aws_security_group.this.id
  description       = "PostgreSQL access"
  cidr_ipv4         = var.postgres_ingress_cidr
  from_port         = var.postgres_port
  to_port           = var.postgres_port
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.this.id
  description       = "Allow all outbound traffic"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

resource "aws_instance" "this" {
  ami                         = data.aws_ssm_parameter.ami.value
  instance_type               = var.instance_type
  subnet_id                   = data.aws_subnet.selected.id
  vpc_security_group_ids      = [aws_security_group.this.id]
  key_name                    = aws_key_pair.this.key_name
  associate_public_ip_address = var.associate_public_ip
  iam_instance_profile        = aws_iam_instance_profile.this.name
  user_data                   = var.user_data

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size_gb
    encrypted             = true
    delete_on_termination = true
  }

  metadata_options {
    http_tokens   = "required"
    http_endpoint = "enabled"
  }

  tags        = merge(var.tags, { Name = var.name })
  volume_tags = merge(var.tags, { Name = "${var.name}-root" })

  # The bootstrap script reads the password from SSM, so the parameter and the
  # role granting access to it must exist before the instance boots.
  depends_on = [aws_ssm_parameter.admin_password, aws_iam_role_policy.this]
}

# Dedicated, persistent volume for the PostgreSQL data directory. It is kept
# separate from the root volume so the instance can be replaced without data loss.
resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.this.availability_zone
  size              = var.data_volume_size_gb
  type              = "gp3"
  encrypted         = true
  tags              = merge(var.tags, { Name = "${var.name}-data" })
}

resource "aws_volume_attachment" "data" {
  device_name = var.data_device_name
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.this.id
}
