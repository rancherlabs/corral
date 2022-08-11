variable "string_output" {
  type = string
}

variable "number_output" {
  type = number
}

variable "array_output" {
  type = list(any)
}

variable "object_output" {
  type = any
}

output "string_output" {
  value = var.string_output
}

output "number_output" {
  value = var.number_output
}

output "array_output" {
  value = var.array_output
}

output "object_output" {
  value = var.object_output
}