# GitHub Copilot Custom Instructions

## Testing Guidelines

- **No temporary test files**: Never create standalone test files like "test-*.dockerfile", "quick-test.*", etc.
- **Use existing test suites**: Add proper test cases to existing test files in the appropriate `*_test.go` files
- **Follow test patterns**: Use the established testing patterns in the codebase (table-driven tests, proper assertions)
- **Clean testing**: If testing is needed, extend the existing test suite rather than creating throwaway files
- **Focus on correct behavior**: Tests should verify expected functionality, not document bugs or regressions
- **Positive test naming**: Name tests to describe what they verify (e.g., `Test_separateScratchConnections`) rather than what's broken
- **Clear test comments**: Comments should explain what behavior is being tested, not reference historical issues

## Code Quality

- Follow the existing code style and patterns
- Use proper error handling
- Write comprehensive tests for new functionality
- Maintain backward compatibility unless explicitly breaking changes are needed
