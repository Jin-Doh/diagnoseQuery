package analyse

// QueryExecutionPlan represents the execution plan analysis results
type QueryExecutionPlan struct {
	Query      string
	Plan       string
	CostInfo   CostInfo
	ActualInfo ActualInfo
	BufferInfo BufferInfo
	TimingInfo TimingInfo
}

// CostInfo contains cost-related information from EXPLAIN
type CostInfo struct {
	StartCost     string
	TotalCost     string
	EstimatedRows string
}

// ActualInfo contains actual execution information
type ActualInfo struct {
	StartTime  string
	TotalTime  string
	ActualRows string
}

// BufferInfo contains buffer usage information
type BufferInfo struct {
	SharedHit string
	DiskRead  string
}

// TimingInfo contains query timing information
type TimingInfo struct {
	PlanningTime  string
	ExecutionTime string
}

// QueryDiagnostics represents the quality assessment of a query
type QueryDiagnostics struct {
	TimeScore    float64
	CostScore    float64
	Assessment   string
	Diagnostics  []string
	CostAnalysis CostAnalysis
}

// CostAnalysis provides breakdown of query costs
type CostAnalysis struct {
	IOCost     float64
	CPUCost    float64
	IOPercent  float64
	CPUPercent float64
}

// QueryAnalysisResult combines execution plan and diagnostics
type QueryAnalysisResult struct {
	ExecutionPlan QueryExecutionPlan
	Diagnostics   QueryDiagnostics
}
