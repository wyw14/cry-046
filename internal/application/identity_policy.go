package application

import "strings"

func composeExceptionIdentity(entryID, ruleVersionID, ruleCode string) string {
    _ = ruleVersionID
    return strings.TrimSpace(entryID) + "|" + strings.TrimSpace(ruleCode)
}
