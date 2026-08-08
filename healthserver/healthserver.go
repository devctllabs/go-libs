package healthserver

import (
	"context"
	"errors"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthserver/internal/generated"
	"github.com/labstack/echo/v5"
)

// Register adds the fixed health probe routes to e.
func Register(e *echo.Echo, probes *health.Probes) error {
	if e == nil {
		return errors.New("healthserver: Echo must not be nil")
	}
	if probes == nil {
		return errors.New("healthserver: Probes must not be nil")
	}

	handler := generated.NewStrictHandler(&strictServer{probes: probes}, nil)
	probeMiddleware := []echo.MiddlewareFunc{probeErrors, noStore}
	generated.RegisterHandlersWithOptions(e, handler, generated.RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			"getLivenessProbe":  probeMiddleware,
			"getReadinessProbe": probeMiddleware,
		},
	})
	return nil
}

type strictServer struct {
	probes *health.Probes
}

func (s *strictServer) GetLivenessProbe(
	_ context.Context,
	_ generated.GetLivenessProbeRequestObject,
) (generated.GetLivenessProbeResponseObject, error) {
	return generated.GetLivenessProbe200JSONResponse{
		ProbeSucceededJSONResponse: probeResponse(s.probes.Liveness()),
	}, nil
}

func (s *strictServer) GetReadinessProbe(
	ctx context.Context,
	request generated.GetReadinessProbeRequestObject,
) (generated.GetReadinessProbeResponseObject, error) {
	report := s.probes.Readiness(ctx)
	response := readinessResponse(report, request.Params.Verbose != nil && *request.Params.Verbose)
	if report.Status == health.StatusOK {
		return generated.GetReadinessProbe200JSONResponse{
			ProbeSucceededJSONResponse: generated.ProbeSucceededJSONResponse(response),
		}, nil
	}
	return generated.GetReadinessProbe503JSONResponse{
		ProbeFailedJSONResponse: generated.ProbeFailedJSONResponse(response),
	}, nil
}

func probeResponse(report health.Report) generated.ProbeSucceededJSONResponse {
	return generated.ProbeSucceededJSONResponse{
		Status: generated.ProbeStatus(report.Status),
	}
}

func readinessResponse(report health.Report, verbose bool) generated.ProbeResponse {
	response := generated.ProbeResponse{Status: generated.ProbeStatus(report.Status)}
	if !verbose {
		return response
	}

	checks := make([]generated.CheckResult, 0, len(report.Checks))
	for _, check := range report.Checks {
		checks = append(checks, generated.CheckResult{
			Name:     check.Name,
			Status:   generated.ProbeStatus(check.Status),
			Critical: check.Critical,
		})
	}
	response.Checks = &checks
	return response
}

func noStore(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		return next(c)
	}
}

func probeErrors(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		var httpError *echo.HTTPError
		if !errors.As(err, &httpError) || httpError.Code != 400 {
			return err
		}

		response := generated.GetReadinessProbe400ApplicationProblemPlusJSONResponse{
			BadRequestApplicationProblemPlusJSONResponse: generated.BadRequestApplicationProblemPlusJSONResponse{
				Type:      generated.ProblemsbadRequest,
				Title:     "Bad Request",
				Status:    generated.N400,
				Retryable: generated.False,
			},
		}
		return response.VisitGetReadinessProbeResponse(c.Response())
	}
}
