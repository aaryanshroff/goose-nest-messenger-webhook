# Source environment variables from .env file
source .env

# Build Go executable
GOOS=linux go build cmd/main.go -o $GO_EXECUTABLE_NAME

# Set up infrastructure and generate deployment package
terraform init
terraform apply

# Remove local binary
rm $GO_EXECUTABLE_NAME

# Remove deployment package
rm $DEPLOYMENT_PACKAGE_NAME