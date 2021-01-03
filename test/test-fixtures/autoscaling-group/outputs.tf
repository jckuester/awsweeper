output "id" {
  value = aws_autoscaling_group.test.id
}

output "id_tag" {
  description = "ASG using the tag attribute for tagging"
  value = aws_autoscaling_group.test_tag.id
}

output "id_tags" {
  description = "ASG using the tags attribute for tagging"
  value = aws_autoscaling_group.test_tags.id
}