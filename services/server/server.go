package server

import (
	"database/sql"

	"context"
	"goprojects/services/generated/auditorpb"
)

type AuditorServer struct {
	auditorpb.UnimplementedClusterAuditorServer
	DB *sql.DB
}

// Dummy values. Logic for calculating health score yet to be implemented.
func (s *AuditorServer) GetHealthScore(ctx context.Context, in *auditorpb.Empty) (*auditorpb.HealthScore, error) {
	return &auditorpb.HealthScore{
		Score:  87.3,
		Status: "Healthy",
	}, nil
}

func (s *AuditorServer) GetFindings(ctx context.Context, in *auditorpb.Empty) (*auditorpb.FindingsResponse, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT namespace, resource, kind, container, issue, suggestion
		FROM findings
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*auditorpb.Finding

	for rows.Next() {
		var f auditorpb.Finding
		err := rows.Scan(&f.Namespace, &f.Resource, &f.Kind, &f.Container, &f.Issue, &f.Suggestion)
		if err != nil {
			return nil, err
		}
		results = append(results, &f)
	}

	return &auditorpb.FindingsResponse{Findings: results}, nil
}
