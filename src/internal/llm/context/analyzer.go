package context

type ContextAnalyzer struct {
    folderPatterns  *PatternMatcher
    userPreferences *PreferenceManager
    workspaceStats  *WorkspaceAnalytics
}

type WorkspaceContext struct {
    CommonPatterns     []FolderPattern
    RecentChanges     []FileOperation
    UserPreferences   Preferences
    WorkspaceMetrics  WorkspaceMetrics
}

func (c *ContextAnalyzer) EnrichDecision(decision DestinationDecision, ctx WorkspaceContext) DestinationDecision {
    // Adjust decision based on workspace context
    if pattern := c.folderPatterns.FindMatchingPattern(decision); pattern != nil {
        decision = pattern.Apply(decision)
    }
    
    if conflict := c.checkConflicts(decision, ctx); conflict != nil {
        decision = c.resolveConflict(conflict)
    }
    
    return c.applyUserPreferences(decision, ctx.UserPreferences)
}