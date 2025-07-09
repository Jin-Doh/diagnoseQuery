package analyse

import (
	"fmt"
	"strings"
)

// ResultFormatter handles formatting of analysis results
type ResultFormatter struct{}

// NewResultFormatter creates a new result formatter
func NewResultFormatter() *ResultFormatter {
	return &ResultFormatter{}
}

// FormatAnalysisResult formats the complete analysis result for display
func (rf *ResultFormatter) FormatAnalysisResult(result *QueryAnalysisResult) string {
	var output strings.Builder

	// Query and execution plan
	output.WriteString(fmt.Sprintf("입력 쿼리: %s\n\n", result.ExecutionPlan.Query))
	output.WriteString("--- 실행 계획 (Execution Plan) ---\n")
	output.WriteString(result.ExecutionPlan.Plan)
	output.WriteString("---------------------------------\n")
	output.WriteString("\n--- 분석 결과 ---\n")

	// Basic information
	rf.writeBasicInfo(&output, result.ExecutionPlan)

	// Quality diagnostics
	rf.writeQualityDiagnostics(&output, result.Diagnostics)

	return output.String()
}

// writeBasicInfo writes basic execution information
func (rf *ResultFormatter) writeBasicInfo(output *strings.Builder, plan QueryExecutionPlan) {
	// Cost information
	if plan.CostInfo.StartCost != "" && plan.CostInfo.TotalCost != "" {
		output.WriteString(fmt.Sprintf("• 예상 비용 (Cost): 시작=%s, 총계=%s\n",
			plan.CostInfo.StartCost, plan.CostInfo.TotalCost))
	}
	if plan.CostInfo.EstimatedRows != "" {
		output.WriteString(fmt.Sprintf("• 예상 행 수 (Rows): %s\n", plan.CostInfo.EstimatedRows))
	}

	// Actual execution information
	if plan.ActualInfo.StartTime != "" && plan.ActualInfo.TotalTime != "" {
		output.WriteString(fmt.Sprintf("• 실제 시간 (Time): 시작=%s ms, 총계=%s ms\n",
			plan.ActualInfo.StartTime, plan.ActualInfo.TotalTime))
	}
	if plan.ActualInfo.ActualRows != "" {
		output.WriteString(fmt.Sprintf("• 실제 행 수 (Rows): %s\n", plan.ActualInfo.ActualRows))
	}

	// Buffer information
	if plan.BufferInfo.SharedHit != "" {
		output.WriteString(fmt.Sprintf("• 버퍼 사용량: 메모리(hit)=%s blocks, 디스크(read)=%s blocks\n",
			plan.BufferInfo.SharedHit, plan.BufferInfo.DiskRead))
	}

	// Timing information
	if plan.TimingInfo.PlanningTime != "" {
		output.WriteString(fmt.Sprintf("• 계획 수립 시간: %s\n", plan.TimingInfo.PlanningTime))
	}
	if plan.TimingInfo.ExecutionTime != "" {
		output.WriteString(fmt.Sprintf("• 쿼리 실행 시간: %s\n", plan.TimingInfo.ExecutionTime))
	}
}

// writeQualityDiagnostics writes quality assessment and recommendations
func (rf *ResultFormatter) writeQualityDiagnostics(output *strings.Builder, diagnostics QueryDiagnostics) {
	output.WriteString("\n--- 쿼리 품질 종합 진단 ---\n")

	// Check for empty diagnostics
	if diagnostics.TimeScore == -1 && diagnostics.CostScore == 10.0 && len(diagnostics.Diagnostics) == 0 {
		output.WriteString("진단 결과: 쿼리 실행 계획을 분석할 수 없습니다. EXPLAIN ANALYZE 결과가 올바른지 확인하세요.\n")
		return
	}

	// Overall assessment and scores
	output.WriteString(fmt.Sprintf("• 종합 평가: %s\n", diagnostics.Assessment))
	output.WriteString(fmt.Sprintf("• 시간 적합도: %.1f / 10.0\n", diagnostics.TimeScore))
	output.WriteString(fmt.Sprintf("• 비용 효율성: %.1f / 10.0\n", diagnostics.CostScore))

	// Cost analysis breakdown
	if diagnostics.CostAnalysis.IOCost+diagnostics.CostAnalysis.CPUCost > 0 {
		output.WriteString(fmt.Sprintf("• 비용 원인 분석: I/O ≈ %.1f%%, CPU ≈ %.1f%%\n",
			diagnostics.CostAnalysis.IOPercent, diagnostics.CostAnalysis.CPUPercent))
	}

	// Detailed diagnostics and recommendations
	if len(diagnostics.Diagnostics) > 0 {
		output.WriteString("\n• 상세 분석 및 개선 제안:\n")
		for i, item := range diagnostics.Diagnostics {
			output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, item))
		}
	} else {
		output.WriteString("\n• 상세 분석 및 개선 제안: 특별한 문제점이 발견되지 않았습니다. 쿼리가 효율적으로 실행되고 있습니다.\n")
	}

	output.WriteString("\n--------------------------------------------------\n")
	output.WriteString("이 기능은 실험적이며, 진단 결과는 참고용으로 활용하시기 바랍니다.\n")
}

// FormatBasicInfo formats only the basic execution plan information
func (rf *ResultFormatter) FormatBasicInfo(plan QueryExecutionPlan) string {
	var output strings.Builder
	rf.writeBasicInfo(&output, plan)
	return output.String()
}

// FormatDiagnosticsOnly formats only the quality diagnostics
func (rf *ResultFormatter) FormatDiagnosticsOnly(diagnostics QueryDiagnostics) string {
	var output strings.Builder
	rf.writeQualityDiagnostics(&output, diagnostics)
	return output.String()
}
