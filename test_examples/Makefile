.PHONY: clean init plan apply

default: init plan apply

clean:
	rm -rf .terraform
	rm -f .terraform.lock.hcl
	rm -f terraform.tfplan
	rm -f terraform.tfstate*

init:
	rm -f .terraform.lock.hcl
	terraform init

plan:
	terraform plan -out terraform.tfplan

apply:
	terraform apply terraform.tfplan

destroy:
	terraform destroy -auto-approve
	make clean
