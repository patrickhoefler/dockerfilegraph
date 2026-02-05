# Agent Instructions

## Testing Guidelines

- **TDD Approach**: Follow Red-Green-Refactor cycle for all new features
  - **Red**: Write failing tests first that describe the desired behavior
  - **Green**: Write minimal code to make tests pass
  - **Refactor**: Clean up code while keeping all tests passing
- **No temporary test files**: Never create standalone test files like "test-*.dockerfile", "quick-test.*", etc.
- **Use existing test suites**: Add proper test cases to existing test files in the appropriate `*_test.go` files
- **Follow test patterns**: Use the established testing patterns in the codebase (table-driven tests, proper assertions)
- **Clean testing**: If testing is needed, extend the existing test suite rather than creating throwaway files
- **Focus on correct behavior**: Tests should verify expected functionality, not document bugs or regressions
- **Test naming**: Name tests to describe what they verify (e.g., `Test_separateScratchConnections`) rather than what's broken
- **Clear test comments**: Comments should explain what behavior is being tested, not reference historical issues
- **Comprehensive coverage**: Ensure edge cases, error conditions, and integration scenarios are tested

## Code Quality

- Follow the existing code style and patterns
- Use proper error handling
- Write comprehensive tests for new functionality
- Maintain backward compatibility unless explicitly breaking changes are needed
- Always use `task check` for efficient linting and testing - it runs golangci-lint, tests, and enforces code quality standards in one command

## Development Tools

- **mise**: Used for tool version management (ensures consistent task version across environments)
- **task**: Task runner for common development tasks
  - Run `task check` before committing changes
  - Use `task build` to create the binary
  - See `task --list` for all available commands
