package scanner

type CmdTokenType int

const (
	ILLEGAL_TOKEN CmdTokenType = iota

	WORD // TODO: rename to FRAGMENT?
	SPACE
	ARG
	// ARG_FLAGS

	SINGLE_QUOTE // '
	DOUBLE_QUOTE // "

	LBRACE_VALUE // ${
	RBRACE       // }
	FORMATTER    // :%s

	LDOUBLE_GUILLEMET // «
	RDOUBLE_GUILLEMET // »
	LSINGLE_GUILLEMET // ‹
	RSINGLE_GUILLEMET // ›

	EOF

	// «\«hello» world\»
	// "\${hello sina" and world \}
)

func (t CmdTokenType) String() string {
	return map[CmdTokenType]string{
		ILLEGAL_TOKEN: "ILLEGAL_TOKEN",

		WORD:  "WORD",
		SPACE: "SPACE",
		ARG:   "ARG",
		// ARG_FLAGS: "ARG_FLAGS",

		SINGLE_QUOTE: "SINGLE_QUOTE",
		DOUBLE_QUOTE: "DOUBLE_QUOTE",

		LBRACE_VALUE: "LBRACE_VALUE",
		RBRACE:       "RBRACE",
		FORMATTER:    "FORMATTER",

		LDOUBLE_GUILLEMET: "LDOUBLE_GUILLEMET",
		RDOUBLE_GUILLEMET: "RDOUBLE_GUILLEMET",
		LSINGLE_GUILLEMET: "LSINGLE_GUILLEMET",
		RSINGLE_GUILLEMET: "RSINGLE_GUILLEMET",

		EOF: "EOF",
	}[t]
}
