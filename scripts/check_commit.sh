#!/bin/bash

# Conventional commit message validator

set -e

# Constants
MAX_DESCRIPTION_LENGTH=70

# Valid commit types (space-separated list)
VALID_TYPES="feat fix docs style refactor perf test chore build ci"

# Function to check if type is valid
is_valid_type() {
    local type="$1"
    for valid_type in $VALID_TYPES; do
        if [[ "$type" == "$valid_type" ]]; then
            return 0
        fi
    done
    return 1
}

# Function to get type description
get_type_description() {
    case "$1" in
        "feat") echo "A new feature" ;;
        "fix") echo "A bug fix" ;;
        "docs") echo "Documentation only changes" ;;
        "style") echo "Changes that do not affect the meaning of the code" ;;
        "refactor") echo "A code change that neither fixes a bug nor adds a feature" ;;
        "perf") echo "A code change that improves performance" ;;
        "test") echo "Adding missing tests or correcting existing tests" ;;
        "chore") echo "Changes to the build process or auxiliary tools" ;;
        "build") echo "Changes that affect the build system or external dependencies" ;;
        "ci") echo "Changes to CI configuration files and scripts" ;;
        *) echo "Unknown type" ;;
    esac
}

# Function to print error messages
print_error() {
    local msg="$1"
    echo "❌ Invalid commit message: $msg" >&2
    echo "" >&2
    echo "Commit message must follow Conventional Commits format:" >&2
    echo "  <type>[optional scope]: <description>" >&2
    echo "" >&2
    echo "Valid types:" >&2

    for type in $VALID_TYPES; do
        printf "  %-10s %s\n" "$type:" "$(get_type_description "$type")" >&2
    done

    echo "" >&2
    echo "Examples:" >&2
    echo "  feat: implement user authentication functionality" >&2
    echo "  fix(database): resolve database connection issue" >&2
    echo "  docs: update readme with installation steps" >&2
    echo "  feat!: drop support for user api version 1" >&2
}

# Function to validate commit message
validate_commit() {
    local commit_msg="$1"

    # Skip empty commits or merge commits
    if [[ -z "$commit_msg" || "$commit_msg" == Merge* ]]; then
        exit 0
    fi

    # Parse conventional commit format using regex
    # Matches: type(scope): description or type: description
    # Also handles breaking changes with
    local regex='^([a-zA-Z]+)(\([^)]+\))?(!)?:[[:space:]]*(.+)$'
    if [[ ! "$commit_msg" =~ $regex ]]; then
        print_error "commit message does not follow conventional commit format"
        exit 1
    fi

    local type="${BASH_REMATCH[1]}"
    local scope="${BASH_REMATCH[2]}"
    local breaking="${BASH_REMATCH[3]}"
    local description="${BASH_REMATCH[4]}"

    # Remove parentheses from scope if present
    if [[ -n "$scope" ]]; then
        scope="${scope#(}"
        scope="${scope%)}"
    fi

    # Trim whitespace from description
    description=$(echo "$description" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    # Validate commit type
    if ! is_valid_type "$type"; then
        print_error "invalid commit type '$type'"
        exit 1
    fi

    # Validate description
    if [[ -z "$description" ]]; then
        print_error "commit description cannot be empty"
        exit 1
    fi

    if [[ ${#description} -gt $MAX_DESCRIPTION_LENGTH ]]; then
        print_error "commit description too long (${#description} chars), should be ≤${MAX_DESCRIPTION_LENGTH} characters"
        exit 1
    fi

    # Check if description ends with period
    if [[ "$description" == *. ]]; then
        print_error "commit description should not end with a period"
        exit 1
    fi

    # Check if description starts with uppercase (should be lowercase)
    local first_char="${description:0:1}"
    if [[ "$first_char" =~ [A-Z] ]]; then
        print_error "commit description should start with lowercase letter"
        exit 1
    fi

    echo "✅ Valid conventional commit: $commit_msg" >&2
}

# Main function
main() {
    if [[ $# -ne 1 ]]; then
        echo "Usage: $0 <commit-msg-file>" >&2
        exit 1
    fi

    local commit_msg_file="$1"

    if [[ ! -f "$commit_msg_file" ]]; then
        echo "Error reading commit message: opening commit message file: $commit_msg_file" >&2
        exit 1
    fi

    # Read the first line of the commit message file
    local commit_msg
    commit_msg=$(head -n 1 "$commit_msg_file" 2>/dev/null | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    if [[ -z "$commit_msg" ]]; then
        echo "Error reading commit message: reading commit message" >&2
        exit 1
    fi

    validate_commit "$commit_msg"
}

# Run main function with all arguments
main "$@"
