package mxidwc

import "testing"

func TestRuleToRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		err      bool
	}{
		{
			name:     "simple pattern without wildcards succeeds",
			input:    "@someone:example.com",
			expected: `^@someone:example\.com$`,
			err:      false,
		},
		{
			name:     "pattern with wildcard as the whole local part succeeds",
			input:    "@*:example.com",
			expected: `^@([^:@]*):example\.com$`,
			err:      false,
		},
		{
			name:     "pattern with wildcard within the local part succeeds",
			input:    "@bot.*.something:example.com",
			expected: `^@bot\.([^:@]*)\.something:example\.com$`,
			err:      false,
		},
		{
			name:     "pattern with wildcard as the whole domain part succeeds",
			input:    "@someone:*",
			expected: `^@someone:([^:@]*)$`,
			err:      false,
		},
		{
			name:     "pattern with wildcard within the domain part succeeds",
			input:    "@someone:*.organization.com",
			expected: `^@someone:([^:@]*)\.organization\.com$`,
			err:      false,
		},
		{
			name:     "pattern with wildcard in both parts succeeds",
			input:    "@*:*",
			expected: `^@([^:@]*):([^:@]*)$`,
			err:      false,
		},
		{
			name:     "pattern that does not appear fully-qualified fails",
			input:    "someone:example.com",
			expected: ``,
			err:      true,
		},
		{
			name:     "pattern that does not appear fully-qualified fails",
			input:    "@someone",
			expected: ``,
			err:      true,
		},
		{
			name:     "pattern with empty domain part fails",
			input:    "@someone:",
			expected: ``,
			err:      true,
		},
		{
			name:     "pattern with empty local part fails",
			input:    "@:example.com",
			expected: ``,
			err:      true,
		},
		{
			name:     "pattern with multiple @ fails",
			input:    "@someone@someone:example.com",
			expected: ``,
			err:      true,
		},
		{
			name:     "pattern with multiple : fails",
			input:    "@someone:someone:example.com",
			expected: ``,
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParsePattern(test.input)

			if test.err {
				if err != nil {
					return
				}
				t.Errorf("expected an error, but did not get one")
			}
			if err != nil {
				t.Errorf("did not expect an error, but got one: %s", err)
			}
			if actual.String() == test.expected {
				return
			}
			t.Errorf(
				"Expected `%s` to yield `%s`, not `%s`",
				test.input,
				test.expected,
				actual.String(),
			)
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		allowedUsers []string
		expected     bool
	}{
		{
			name:         "Empty allowed users allows no one",
			input:        "@someone:example.com",
			allowedUsers: []string{},
			expected:     false,
		},
		{
			name:         "Direct full mxid match is allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@someone:example.com"},
			expected:     true,
		},
		{
			name:         "Direct full mxid match later on is allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@another:example.com", "@someone:example.com"},
			expected:     true,
		},
		{
			name:         "No mxid match is not allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@another:example.com"},
			expected:     false,
		},
		{
			name:         "mxid localpart only wildcard match is allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@*:example.com"},
			expected:     true,
		},
		{
			name:         "mxid localpart with wildcard match is allowed",
			input:        "@bot.abc:example.com",
			allowedUsers: []string{"@bot.*:example.com"},
			expected:     true,
		},
		{
			name:         "mxid localpart with wildcard match is not allowed when it does not match",
			input:        "@bot.abc:example.com",
			allowedUsers: []string{"@employee.*:example.com"},
			expected:     false,
		},
		{
			name:         "mxid localpart wildcard for another domain is not allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@*:another.com"},
			expected:     false,
		},
		{
			name:         "mxid domainpart with only wildcard match is allowed",
			input:        "@someone:example.com",
			allowedUsers: []string{"@someone:*"},
			expected:     true,
		},
		{
			name:         "mxid domainpart with wildcard match is allowed",
			input:        "@someone:example.organization.com",
			allowedUsers: []string{"@someone:*.organization.com"},
			expected:     true,
		},
		{
			name:         "mxid domainpart with wildcard match is not allowed when it does not match",
			input:        "@someone:example.another.com",
			allowedUsers: []string{"@someone:*.organization.com"},
			expected:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regexes, err := ParsePatterns(test.allowedUsers)
			if err != nil {
				t.Error(err)
			}

			actual := Match(test.input, regexes)

			if actual == test.expected {
				return
			}
			t.Errorf(
				"Expected `%s` compared against `%v` to yield `%v`, not `%v`",
				test.input,
				test.allowedUsers,
				test.expected,
				actual,
			)
		})
	}
}
