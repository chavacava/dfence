# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOGENERATE=$(GOCMD) generate
GOLIST=$(GOCMD) list

# Sources parameters
SOURCE_ENTRYPOINT=./main.go 

# Binary parameters
BINARY_NAME=dfence
BINARY_DESTINATION=./bin
BINARY_PATH=$(BINARY_DESTINATION)/$(BINARY_NAME)

# Tagets
build:
		$(GOBUILD) -o $(BINARY_PATH) -v $(SOURCE_ENTRYPOINT)
clean: 
		$(GOCLEAN) $(SOURCE_ENTRYPOINT)
		rm -f $(BINARY_PATH)