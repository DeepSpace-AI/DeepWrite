package agent

import "strings"

var DefaultSafetyRules = []string{
	"Protect user privacy and confidentiality at all times",
	"Never reveal sensitive personal information such as passwords, API keys, or private keys",
	"Provide accurate and helpful information to the best of your abilities",
	"Acknowledge uncertainty when you don't know something",
	"Respect user boundaries and preferences",
	"Follow ethical AI guidelines and avoid harmful content",
	"Do not generate content that promotes illegal activities",
	"Be transparent about your capabilities and limitations",
}

func GetDefaultSafetyPrompt() string {
	return buildSafetyPrompt(DefaultSafetyRules)
}

func GetDefaultSafetyPromptWithRules(rules []string) string {
	if len(rules) == 0 {
		return GetDefaultSafetyPrompt()
	}
	return buildSafetyPrompt(rules)
}

func buildSafetyPrompt(rules []string) string {
	var sb strings.Builder

	sb.WriteString("# Safety Guidelines\n\n")
	sb.WriteString("Follow these guidelines to ensure safe and helpful interactions:\n\n")

	for i, rule := range rules {
		sb.WriteString("- ")
		sb.WriteString(rule)
		if i < len(rules)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func GetMinimalSafetyPrompt() string {
	return `# Safety Guidelines

- Protect user privacy and confidentiality
- Do not reveal sensitive information
- Provide accurate and helpful responses
- Acknowledge limitations when uncertain`
}
