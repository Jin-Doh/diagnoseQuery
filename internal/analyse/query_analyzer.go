package analyse

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// QueryAnalyzer handles PostgreSQL query analysis
type QueryAnalyzer struct {
	db *sql.DB
}

// NewQueryAnalyzer creates a new query analyzer instance
func NewQueryAnalyzer(db *sql.DB) *QueryAnalyzer {
	return &QueryAnalyzer{db: db}
}

// AnalyzeQuery performs comprehensive analysis of a SQL query
func (qa *QueryAnalyzer) AnalyzeQuery(query string) (QueryAnalysisResult, error) {
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), "select") {
		return QueryAnalysisResult{}, fmt.Errorf("비용 분석은 SELECT 쿼리만 지원합니다")
	}

	// Execute EXPLAIN ANALYZE
	plan, err := qa.getExecutionPlan(query)
	if err != nil {
		return QueryAnalysisResult{}, fmt.Errorf("실행 계획 조회 실패: %v", err)
	}

	// Parse execution plan
	executionPlan := qa.parseExecutionPlan(query, plan)

	// Generate diagnostics
	diagnostics := qa.generateDiagnostics(query, plan, executionPlan)

	return QueryAnalysisResult{
		ExecutionPlan: executionPlan,
		Diagnostics:   diagnostics,
	}, nil
}

// getExecutionPlan executes EXPLAIN (ANALYZE, BUFFERS) and returns the plan
func (qa *QueryAnalyzer) getExecutionPlan(query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS) %s", query)

	rows, err := qa.db.Query(explainQuery)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var planBuilder strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		planBuilder.WriteString(line)
		planBuilder.WriteString("\n")
	}

	return planBuilder.String(), nil
}

// parseExecutionPlan extracts structured information from the execution plan
func (qa *QueryAnalyzer) parseExecutionPlan(query, plan string) QueryExecutionPlan {
	firstLine := strings.Split(plan, "\n")[0]

	// Define regex patterns
	costRe := regexp.MustCompile(`cost=([\d\.]+)\.\.([\d\.]+) rows=([\d]+)`)
	actualRe := regexp.MustCompile(`actual time=([\d\.]+)\.\.([\d\.]+) rows=([\d]+)`)
	buffersRe := regexp.MustCompile(`Buffers: shared hit=(\d+)(?: read=(\d+))?`)
	planningTimeRe := regexp.MustCompile(`Planning Time: ([\d\.]+ ms)`)
	executionTimeRe := regexp.MustCompile(`Execution Time: ([\d\.]+ ms)`)

	// Extract matches
	costMatches := costRe.FindStringSubmatch(firstLine)
	actualMatches := actualRe.FindStringSubmatch(firstLine)
	buffersMatches := buffersRe.FindStringSubmatch(plan)
	planningMatch := planningTimeRe.FindStringSubmatch(plan)
	executionMatch := executionTimeRe.FindStringSubmatch(plan)

	executionPlan := QueryExecutionPlan{
		Query: query,
		Plan:  plan,
	}

	// Parse cost information
	if len(costMatches) >= 4 {
		executionPlan.CostInfo = CostInfo{
			StartCost:     costMatches[1],
			TotalCost:     costMatches[2],
			EstimatedRows: costMatches[3],
		}
	}

	// Parse actual execution information
	if len(actualMatches) >= 4 {
		executionPlan.ActualInfo = ActualInfo{
			StartTime:  actualMatches[1],
			TotalTime:  actualMatches[2],
			ActualRows: actualMatches[3],
		}
	}

	// Parse buffer information
	if len(buffersMatches) > 1 {
		hit := buffersMatches[1]
		read := "0"
		if len(buffersMatches) > 2 && buffersMatches[2] != "" {
			read = buffersMatches[2]
		}
		executionPlan.BufferInfo = BufferInfo{
			SharedHit: hit,
			DiskRead:  read,
		}
	}

	// Parse timing information
	if len(planningMatch) > 1 {
		executionPlan.TimingInfo.PlanningTime = planningMatch[1]
	}
	if len(executionMatch) > 1 {
		executionPlan.TimingInfo.ExecutionTime = executionMatch[1]
	}

	return executionPlan
}

// generateDiagnostics creates quality assessment and recommendations
func (qa *QueryAnalyzer) generateDiagnostics(query, plan string, executionPlan QueryExecutionPlan) QueryDiagnostics {
	diagnostics := QueryDiagnostics{
		TimeScore: -1.0,
		CostScore: 10.0,
	}

	var issues []string

	// Calculate time score
	if executionPlan.TimingInfo.ExecutionTime != "" {
		execTimeStr := strings.TrimSuffix(executionPlan.TimingInfo.ExecutionTime, " ms")
		if execTime, err := strconv.ParseFloat(execTimeStr, 64); err == nil {
			diagnostics.TimeScore = qa.calculateTimeScore(execTime)
			if diagnostics.TimeScore < 5 {
				issues = append(issues, fmt.Sprintf("실행 시간이 %.2fms로, 설정된 기준보다 깁니다. 쿼리 로직 또는 인덱스 사용을 최적화하여 응답 시간을 단축해야 합니다.", execTime))
			}
		}
	}

	// Calculate cost score and analyze issues
	costIssues := qa.analyzeCostEfficiency(query, plan, executionPlan)
	issues = append(issues, costIssues...)

	// Calculate cost penalties
	penalties := qa.calculateCostPenalties(query, plan, executionPlan)
	diagnostics.CostScore -= penalties

	if diagnostics.CostScore < 0 {
		diagnostics.CostScore = 0
	}

	// Calculate cost analysis (I/O vs CPU)
	diagnostics.CostAnalysis = qa.calculateCostAnalysis(executionPlan)

	// Determine overall assessment
	diagnostics.Assessment = qa.determineOverallAssessment(diagnostics.TimeScore, diagnostics.CostScore)
	diagnostics.Diagnostics = issues

	return diagnostics
}

// calculateTimeScore assigns a score based on execution time
func (qa *QueryAnalyzer) calculateTimeScore(execTime float64) float64 {
	switch {
	case execTime < 10:
		return 10
	case execTime < 50:
		return 9
	case execTime < 100:
		return 8
	case execTime < 250:
		return 7
	case execTime < 500:
		return 6
	case execTime < 1000:
		return 5
	case execTime < 2000:
		return 4
	case execTime < 5000:
		return 3
	case execTime < 10000:
		return 2
	default:
		return 1
	}
}

// analyzeCostEfficiency analyzes cost-related issues and returns recommendations
func (qa *QueryAnalyzer) analyzeCostEfficiency(query, plan string, executionPlan QueryExecutionPlan) []string {
	var issues []string

	// Check for Sequential Scan issues
	seqScanIssues := qa.analyzeSequentialScan(query, plan)
	issues = append(issues, seqScanIssues...)

	// Check for estimated vs actual row differences
	rowDiffIssue := qa.analyzeRowDifferences(executionPlan)
	if rowDiffIssue != "" {
		issues = append(issues, rowDiffIssue)
	}

	// Check for excessive disk reads
	diskReadIssue := qa.analyzeDiskReads(executionPlan)
	if diskReadIssue != "" {
		issues = append(issues, diskReadIssue)
	}

	return issues
}

// analyzeSequentialScan checks for sequential scan issues
func (qa *QueryAnalyzer) analyzeSequentialScan(query, plan string) []string {
	var issues []string

	if !strings.Contains(plan, "Seq Scan") {
		return issues
	}

	upperQuery := strings.ToUpper(query)
	if !strings.Contains(upperQuery, "WHERE") && !strings.Contains(upperQuery, "JOIN") {
		issues = append(issues, "쿼리에 WHERE 또는 JOIN 절이 없어 전체 테이블을 스캔하고 있습니다. 불필요한 데이터 스캔을 줄이기 위해 필터 조건을 추가하는 것을 강력히 권장합니다.")
	} else {
		tableIssue := qa.analyzeSequentialScanTables(plan)
		if tableIssue != "" {
			issues = append(issues, tableIssue)
		}
	}

	// Check for large sequential scans
	largeSeqScanIssue := qa.analyzeLargeSequentialScan(plan)
	if largeSeqScanIssue != "" {
		issues = append(issues, largeSeqScanIssue)
	}

	return issues
}

// analyzeSequentialScanTables identifies tables with sequential scans
func (qa *QueryAnalyzer) analyzeSequentialScanTables(plan string) string {
	seqScanRe := regexp.MustCompile(`Seq Scan on ([\w_]+)`)
	matches := seqScanRe.FindAllStringSubmatch(plan, -1)

	tables := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	if len(tables) > 0 {
		return fmt.Sprintf("비효율적인 'Sequential Scan'이 '%s' 테이블에서 발견되었습니다. WHERE 절이나 JOIN 조건에 사용된 칼럼에 인덱스 생성을 고려해 보세요.", strings.Join(tables, ", "))
	}

	return "비효율적인 'Sequential Scan'이 발견되었습니다. 인덱스를 활용하도록 쿼리를 수정하는 것을 고려해 보세요."
}

// analyzeLargeSequentialScan checks for sequential scans with many rows
func (qa *QueryAnalyzer) analyzeLargeSequentialScan(plan string) string {
	seqScanRowsRe := regexp.MustCompile(`Seq Scan on .* rows=([\d]+)`)
	seqScanMatches := seqScanRowsRe.FindStringSubmatch(plan)

	if len(seqScanMatches) > 1 {
		if rows, _ := strconv.Atoi(seqScanMatches[1]); rows > 10000 {
			return fmt.Sprintf("Sequential Scan으로 처리하는 행의 수가 %d개로 매우 많습니다. 이는 성능 저하의 주된 원인일 수 있습니다.", rows)
		}
	}

	return ""
}

// analyzeRowDifferences checks for differences between estimated and actual rows
func (qa *QueryAnalyzer) analyzeRowDifferences(executionPlan QueryExecutionPlan) string {
	if executionPlan.CostInfo.EstimatedRows == "" || executionPlan.ActualInfo.ActualRows == "" {
		return ""
	}

	expected, _ := strconv.ParseFloat(executionPlan.CostInfo.EstimatedRows, 64)
	actual, _ := strconv.ParseFloat(executionPlan.ActualInfo.ActualRows, 64)

	if expected > 0 && actual > 100 && (actual > expected*10 || actual < expected*0.1) {
		return fmt.Sprintf("계획 수립기가 예측한 행 수(%.0f)와 실제 반환된 행 수(%.0f)의 차이가 큽니다. 이는 최적의 실행 계획을 세우지 못하게 합니다. 'ANALYZE [테이블명];'을 실행하여 통계 정보를 갱신하세요.", expected, actual)
	}

	return ""
}

// analyzeDiskReads checks for excessive disk reads
func (qa *QueryAnalyzer) analyzeDiskReads(executionPlan QueryExecutionPlan) string {
	if executionPlan.BufferInfo.DiskRead == "" || executionPlan.BufferInfo.DiskRead == "0" {
		return ""
	}

	if read, _ := strconv.Atoi(executionPlan.BufferInfo.DiskRead); read > 10 {
		return fmt.Sprintf("디스크에서 읽어온 데이터 페이지가 %d개로 많습니다. I/O 비용을 줄이기 위해 인덱스를 활용하거나, 메모리(shared_buffers) 증설을 고려해 보세요.", read)
	}

	return ""
}

// calculateCostPenalties calculates penalty points for cost score
func (qa *QueryAnalyzer) calculateCostPenalties(query, plan string, executionPlan QueryExecutionPlan) float64 {
	var penalties float64

	// Sequential Scan penalties
	penalties += qa.calculateSequentialScanPenalties(query, plan)

	// Estimated vs actual row difference penalty
	penalties += qa.calculateRowDifferencePenalty(executionPlan)

	// Disk read penalties
	penalties += qa.calculateDiskReadPenalties(executionPlan)

	return penalties
}

// calculateSequentialScanPenalties calculates penalties for sequential scans
func (qa *QueryAnalyzer) calculateSequentialScanPenalties(query, plan string) float64 {
	if !strings.Contains(plan, "Seq Scan") {
		return 0
	}

	var penalties float64
	upperQuery := strings.ToUpper(query)

	if !strings.Contains(upperQuery, "WHERE") && !strings.Contains(upperQuery, "JOIN") {
		penalties += 5 // Higher penalty for queries without filters
	} else {
		penalties += 3
	}

	// Additional penalty for large sequential scans
	seqScanRowsRe := regexp.MustCompile(`Seq Scan on .* rows=([\d]+)`)
	seqScanMatches := seqScanRowsRe.FindStringSubmatch(plan)
	if len(seqScanMatches) > 1 {
		if rows, _ := strconv.Atoi(seqScanMatches[1]); rows > 10000 {
			penalties += 2
		}
	}

	return penalties
}

// calculateRowDifferencePenalty calculates penalty for row estimation differences
func (qa *QueryAnalyzer) calculateRowDifferencePenalty(executionPlan QueryExecutionPlan) float64 {
	if executionPlan.CostInfo.EstimatedRows == "" || executionPlan.ActualInfo.ActualRows == "" {
		return 0
	}

	expected, _ := strconv.ParseFloat(executionPlan.CostInfo.EstimatedRows, 64)
	actual, _ := strconv.ParseFloat(executionPlan.ActualInfo.ActualRows, 64)

	if expected > 0 && actual > 100 && (actual > expected*10 || actual < expected*0.1) {
		return 2
	}

	return 0
}

// calculateDiskReadPenalties calculates penalties for disk reads
func (qa *QueryAnalyzer) calculateDiskReadPenalties(executionPlan QueryExecutionPlan) float64 {
	if executionPlan.BufferInfo.DiskRead == "" || executionPlan.BufferInfo.DiskRead == "0" {
		return 0
	}

	read, _ := strconv.Atoi(executionPlan.BufferInfo.DiskRead)
	if read <= 0 {
		return 0
	}

	switch {
	case read > 1000:
		return 4
	case read > 100:
		return 2
	case read > 10:
		return 1
	default:
		return 0
	}
}

// calculateCostAnalysis breaks down costs into I/O and CPU components
func (qa *QueryAnalyzer) calculateCostAnalysis(executionPlan QueryExecutionPlan) CostAnalysis {
	const seqPageCost = 1.0
	const randomPageCost = 4.0
	const cpuTupleCost = 0.01

	var ioCost, cpuCost float64

	// Calculate I/O cost from buffer info
	if executionPlan.BufferInfo.SharedHit != "" {
		hit, _ := strconv.ParseFloat(executionPlan.BufferInfo.SharedHit, 64)
		ioCost += hit * seqPageCost
	}

	if executionPlan.BufferInfo.DiskRead != "" && executionPlan.BufferInfo.DiskRead != "0" {
		read, _ := strconv.ParseFloat(executionPlan.BufferInfo.DiskRead, 64)
		ioCost += read * randomPageCost
	}

	// Calculate CPU cost from row processing
	if executionPlan.ActualInfo.ActualRows != "" {
		rows, _ := strconv.ParseFloat(executionPlan.ActualInfo.ActualRows, 64)
		cpuCost = rows * cpuTupleCost
	}

	totalCost := ioCost + cpuCost
	var ioPercent, cpuPercent float64

	if totalCost > 0 {
		ioPercent = (ioCost / totalCost) * 100
		cpuPercent = (cpuCost / totalCost) * 100
	}

	return CostAnalysis{
		IOCost:     ioCost,
		CPUCost:    cpuCost,
		IOPercent:  ioPercent,
		CPUPercent: cpuPercent,
	}
}

// determineOverallAssessment provides an overall quality assessment
func (qa *QueryAnalyzer) determineOverallAssessment(timeScore, costScore float64) string {
	if timeScore >= 8.0 && costScore >= 8.0 {
		return "✅ 훌륭함 (추가 최적화 불필요)"
	} else if timeScore < 5.0 || costScore < 5.0 {
		return "⚠️ 개선 필요 (낮은 품질)"
	} else {
		return "💡 개선 고려 (보통 품질)"
	}
}
