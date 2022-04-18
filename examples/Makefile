.PHONY: clean init plan apply

clean:
	rm -rf .terraform
	rm -f .terraform.lock.hcl

init: clean
	terraform init

plan:
	terraform plan -out terraform.tfplan

apply:
	terraform apply terraform.tfplan
