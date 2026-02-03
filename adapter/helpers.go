package adapter

func canonicalDescription(ct *CanonicalTool) string {
	if ct == nil {
		return ""
	}
	if ct.Description != "" {
		return ct.Description
	}
	if ct.Summary != "" {
		return ct.Summary
	}
	if ct.DisplayName != "" {
		return ct.DisplayName
	}
	return ""
}
