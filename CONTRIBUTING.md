## Contributing to bmw-cardata

Thank you for your interest in improving this project! Contributions of all kinds are welcome — bug reports, documentation, tests, and code changes.

This project is licensed under Apache License 2.0. By contributing, you agree to the guidelines below.

### Ways to contribute
- **Report bugs**: Open an issue with steps to reproduce, expected vs. actual behavior, logs, and environment details.
- **Propose features**: Open an issue describing the use case, API sketch, and alternatives considered.
- **Improve docs**: Fix typos, clarify behavior, add examples.
- **Submit code**: Open a pull request following the workflow below.

### Development setup
- **Go version**: Use the version in `go.mod` (currently Go 1.24.5).
- **Clone**: Fork the repo and create a feature branch from `bmw-cardata`.
- **Install deps**: `go mod download`
- **Format**: `gofmt -s -w .`
- **Vet (recommended)**: `go vet ./...`
- **Test**: `go test ./...`

### Coding guidelines
- Keep changes focused and reasonably small; split up large changes.
- Add or update tests for new behavior and edge cases.
- Maintain clear, descriptive names and avoid unnecessary abstractions.
- Do not introduce breaking changes without prior discussion.

### Pull request checklist
- [ ] Linked issue or clear problem statement in the PR description
- [ ] Tests added/updated and `go test ./...` passes locally
- [ ] Public APIs documented and README updated when applicable
- [ ] Code formatted (`gofmt -s -w .`) and vetted (`go vet ./...`)
- [ ] Commits are clear; rebase onto latest `bmw-cardata` before merging

### Communication and conduct
- Be respectful and constructive. Disagreement is fine; disrespect is not.
- Prefer discussion in issues and PRs for transparency and shared context.

### Security
If you believe you have found a security issue, please do not open a public issue. Instead, contact the maintainer privately if possible. Responsible disclosure is appreciated.

### License of this project
This repository is licensed under the Apache License, Version 2.0. See `LICENSE` for details.

### Copyright assignment and license of contributions
By submitting any contribution (including code, documentation, or other materials) to this repository, you agree that:

1. You **irrevocably assign and transfer** to the project maintainer all right, title, and interest worldwide in and to your contribution, including all copyrights and related rights, to the maximum extent permitted by applicable law. If and to the extent any such assignment is ineffective, you grant the project maintainer a **perpetual, worldwide, non‑exclusive, royalty‑free, irrevocable license** to use, reproduce, modify, distribute, sublicense, and otherwise exploit your contribution under the project’s license terms.
2. You agree that your contribution will be distributed by the project under the **Apache License 2.0**, without additional terms or conditions.
3. To the extent permitted by law, you **waive moral rights** (e.g., rights of attribution or integrity) or agree not to assert them against the project maintainer and downstream users.
4. You confirm you have the necessary rights to make the contribution and to grant the rights above; if your employer or another entity owns the rights, you confirm you have their permission.
5. Contributions are done on a voluntary basis and does not open right for any compensation or indeminties.

If your organization requires a separate Contributor License Agreement (CLA) or cannot agree to the assignment above, please open an issue before submitting a pull request to discuss alternatives.

### Sign-off (optional but recommended)
If you prefer using a Developer Certificate of Origin (DCO) sign-off, add a `Signed-off-by: Your Name <email@example.com>` line to your commits certifying that you have the right to submit the work under the stated license and the terms above.

---

Thank you again for contributing!
