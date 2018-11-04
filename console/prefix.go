package console

var (
	prefixCtx = ""
	prefixPmt = " "
)

func Prefix() (string, bool) {
	return "jira>" + prefixCtx + prefixPmt, true
}
