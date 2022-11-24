# Source environment variables from .env file
source .env

# Build Go executable and output it to $GO_EXECUTABLE_NAME
GOOS=linux go build -o $GO_EXECUTABLE_NAME cmd/main.go

# Set up infrastructure and generate deployment package
terraform init
terraform apply

# Remove local binary
rm $GO_EXECUTABLE_NAME

# Remove deployment package
rm $DEPLOYMENT_PACKAGE_NAME