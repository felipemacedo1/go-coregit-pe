# Security Policy

## Overview

Go Core Git takes security seriously. This document outlines our security practices and how to report security vulnerabilities.

## Security Principles

### 1. No Credential Storage
- **Never store credentials**: All authentication is delegated to Git's native credential helpers
- **No credential logging**: URLs with credentials are sanitized in logs
- **Environment isolation**: Minimal environment variables for Git execution

### 2. Input Sanitization
- **Command injection prevention**: All Git command arguments are sanitized
- **Path validation**: Repository paths are validated to prevent traversal attacks
- **Output sanitization**: Sensitive information is redacted from command output

### 3. Secure Execution
- **Controlled environment**: Git commands run with minimal, controlled environment
- **Timeout protection**: All operations have configurable timeouts
- **Process isolation**: Each Git command runs in isolated process context

## Security Features

### Argument Sanitization
```go
// Dangerous characters are filtered from arguments
sanitizedArgs := sanitizeArgs(userInput)
```

### Credential Redaction
```go
// URLs with credentials are sanitized in logs
sanitizedURL := sanitizeURL("https://user:token@github.com/repo.git")
// Result: "https://***@github.com/repo.git"
```

### Environment Control
```go
// Minimal, secure environment for Git execution
cmd.Env = []string{
    "GIT_TERMINAL_PROMPT=0",  // Disable interactive prompts
    "LC_ALL=C",               // Consistent locale
    "PATH=" + getSecurePath(), // Controlled PATH
}
```

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x:                |
| < 0.1   | :x:                |

## Reporting a Vulnerability

### How to Report
1. **Do NOT create a public issue** for security vulnerabilities
2. Send an email to: contato.dev.macedo@gmail.com
3. Include detailed information about the vulnerability
4. Provide steps to reproduce if possible

### What to Include
- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if known)
- Your contact information

### Response Timeline
- **Initial Response**: Within 48 hours
- **Assessment**: Within 1 week
- **Fix Development**: Depends on severity
- **Public Disclosure**: After fix is released

## Security Best Practices for Users

### 1. Credential Management
- Use Git's credential helpers (credential-manager, keychain, etc.)
- Never include credentials in URLs or configuration files
- Use SSH keys with passphrases when possible
- Regularly rotate access tokens

### 2. Network Security
- Use HTTPS for remote repositories when possible
- Verify SSL certificates
- Be cautious with self-signed certificates
- Use VPN for sensitive repositories

### 3. Local Security
- Protect your local Git configuration
- Use file system permissions appropriately
- Keep your Git binary updated
- Regularly audit repository access

### 4. API Security (gitmgr-server)
- Run API server on localhost only (127.0.0.1)
- Use firewall rules to restrict access
- Monitor API access logs
- Consider authentication for production use

## Known Security Considerations

### 1. Local API Server
- **Current State**: No authentication required
- **Risk**: Local processes can access API
- **Mitigation**: Bind to localhost only, use firewall rules
- **Future**: Authentication may be added in future versions

### 2. Git Binary Dependency
- **Risk**: Vulnerable Git binary affects security
- **Mitigation**: Keep Git updated, validate Git version
- **Recommendation**: Use latest stable Git version

### 3. File System Access
- **Risk**: Repository paths could be manipulated
- **Mitigation**: Path validation and sanitization
- **Recommendation**: Run with minimal file system permissions

## Security Testing

### Automated Testing
- Input sanitization tests
- Credential redaction tests
- Path validation tests
- Command injection prevention tests

### Manual Testing
- Penetration testing of API endpoints
- Credential handling verification
- Error message analysis for information leakage

## Security Updates

Security updates will be:
1. Developed privately
2. Tested thoroughly
3. Released as patch versions
4. Announced with security advisory
5. Documented in CHANGELOG.md

## Compliance

This project follows:
- OWASP secure coding practices
- Go security best practices
- Git security recommendations
- Industry standard vulnerability disclosure

## Contact

For security-related questions or concerns:
- Email: contato.dev.macedo@gmail.com
- GitHub: https://github.com/felipemacedo1
- LinkedIn: https://linkedin.com/in/felipemacedo1

## Acknowledgments

We appreciate security researchers who responsibly disclose vulnerabilities. Contributors will be acknowledged in our security advisories (with permission).

---

**Note**: This security policy is a living document and will be updated as the project evolves.