package domain

type Source string

const (
	SourceWebForm Source = "webform"
	SourceManual  Source = "manual"
	SourceAPI     Source = "api"
)

type LeadStatus string

const (
	LeadStatusNew          LeadStatus = "new"
	LeadStatusContacted    LeadStatus = "contacted"
	LeadStatusQualified    LeadStatus = "qualified"
	LeadStatusConverted    LeadStatus = "converted"
	LeadStatusDisqualified LeadStatus = "disqualified"
)

type LeadStage string

const (
	StageCapture LeadStage = "capture"
	StageTriage  LeadStage = "triage"
	StageQualify LeadStage = "qualify"
	StageNurture LeadStage = "nurture"
	StageConvert LeadStage = "convert"
)

type ActivityType string

const (
	ActivityTypeNote                ActivityType = "note"
	ActivityTypeCall                ActivityType = "call"
	ActivityTypeEmail               ActivityType = "email"
	ActivityTypeStatusChange        ActivityType = "status_change"
	ActivityTypeScoreChanged        ActivityType = "score_changed"
	ActivityTypeStageChanged        ActivityType = "stage_changed"
	ActivityTypeEngagementRequested ActivityType = "engagement_request"
	ActivityTypeConversion          ActivityType = "conversion"
)
