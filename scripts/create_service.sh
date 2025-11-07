#!/bin/bash

# Service structure generator

set -e

# Function to show usage
show_usage() {
    echo "Usage: $0 -name <service-name>" >&2
    echo "Example: $0 -name user" >&2
    exit 1
}

# Function to create service structure
create_service_structure() {
    local service_name="$1"
    local base_path="services/${service_name}-service"

    # Create directory structure
    local dirs=(
        "cmd"
        "internal/config"
        "internal/domain"
        "internal/handler"
        "internal/repository"
        "internal/usecase"
        "pkg/client"
        "pkg/types"
    )

    echo "Creating service structure for ${service_name}-service..."

    for dir in "${dirs[@]}"; do
        local full_path="${base_path}/${dir}"
        if ! mkdir -p "$full_path"; then
            echo "Failed to create directory $full_path" >&2
            exit 1
        fi
    done

    # Create README.md content
    local readme_content="# ${service_name} service

This service handles all ${service_name}-related operations in the system.

## Architecture

The service follows Clean Architecture principles with the following structure:

\`\`\`
services/${service_name}-service/
├── cmd/                   # Application entry points
│   └── main.go            # Main application bootstrapper
├── internal/              # Private application code
│   ├── config/            # Application configuration
│   ├── domain/            # Business domain models and interfaces
│   ├── handler/           # Request handlers
│   ├── repository/        # Data persistence layer
│   └── usercase/          # Business logic implementation
├── pkg/                   # Public API and shared code
│   ├── client/            # Client libraries
│   └── types/             # Shared type definitions
└── README.md              # Service documentation
\`\`\`

### Layer Responsibilities

1. **Presentation Layer** (\`internal/handler/\`)
   - Handles incoming requests and responses
   - Translates between transport payloads and domain models
   - Input validation and error handling
   - Protocol-specific implementations (gRPC, HTTP, etc.)

2. **Domain Layer** (\`internal/domain/\`)
   - Contains business domain models and entities
   - Defines contracts for stores and services
   - Pure business logic with no external dependencies
   - Domain-specific errors and validation rules

3. **Application Layer** (\`internal/usecase/\`)
   - Implements business use cases and workflows
   - Orchestrates domain objects and store operations
   - Handles cross-cutting concerns (logging, metrics)
   - Enforces business rules and policies

4. **Data Layer** (\`internal/repository/\`)
   - Handles data persistence and retrieval
   - Implements domain store interfaces
   - Manages database connections and transactions
   - Data mapping and query optimization

## Key Benefits

1. **Dependency Inversion**: Services depend on interfaces, not concrete implementations
2. **Separation of Concerns**: Each layer has a single, well-defined responsibility
3. **Testability**: Easy to mock dependencies and write comprehensive unit tests
4. **Maintainability**: Clear boundaries and contracts between components
5. **Flexibility**: Easy to swap implementations without affecting business logic
6. **Scalability**: Modular design supports independent scaling and deployment

### Configuration
The service uses environment variables for configuration. See \`internal/config/\` for available options.
"

    # Write README.md file
    local readme_path="${base_path}/README.md"
    if ! echo "$readme_content" > "$readme_path"; then
        echo "Failed to create README.md" >&2
        exit 1
    fi

    # Create cmd/main.go with empty main function
    local main_go_content="package main

func main() {
	// TODO: Implement service initialization and startup logic
}
"

    local main_go_path="${base_path}/cmd/main.go"
    if ! echo "$main_go_content" > "$main_go_path"; then
        echo "Failed to create cmd/main.go" >&2
        exit 1
    fi

    echo "Successfully created ${service_name}-service structure in $base_path"
    echo ""
    echo "Directory structure created:"
    echo "services/${service_name}-service/"
    echo "├── cmd/                   # Application entry points"
    echo "│   └── main.go            # Main application bootstrapper"
    echo "├── internal/              # Private application code"
    echo "│   ├── config/            # Application configuration"
    echo "│   ├── domain/            # Business domain models and interfaces"
    echo "│   ├── handler/           # Request handlers"
    echo "│   ├── repository/        # Data persistence layer"
    echo "│   └── usecase/           # Business logic implementation"
    echo "├── pkg/                   # Public API and shared code"
    echo "│   ├── client/            # Client libraries"
    echo "│   └── types/             # Shared type definitions"
    echo "└── README.md              # Service documentation"
}

# Main function
main() {
    local service_name=""

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -name)
                service_name="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                ;;
            *)
                echo "Unknown option: $1" >&2
                show_usage
                ;;
        esac
    done

    # Validate service name
    if [[ -z "$service_name" ]]; then
        echo "Please provide a service name using -name flag" >&2
        show_usage
    fi

    # Validate service name format (basic validation)
    if [[ ! "$service_name" =~ ^[a-zA-Z][a-zA-Z0-9_-]*$ ]]; then
        echo "Invalid service name. Use only letters, numbers, hyphens, and underscores. Must start with a letter." >&2
        exit 1
    fi

    create_service_structure "$service_name"
}

# Run main function with all arguments
main "$@"
