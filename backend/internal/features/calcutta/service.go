package calcutta

import appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"

type Service = appcalcutta.Service

type Ports = appcalcutta.Ports

func New(ports Ports) *Service {
	return appcalcutta.New(ports)
}
