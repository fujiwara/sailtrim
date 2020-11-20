package sailtrim

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/pkg/errors"
)

type SailTrim struct {
	svc  *lightsail.Lightsail
	conf *Config
}

func New(sess *session.Session, path string) (*SailTrim, error) {
	c, err := loadConfig(path)
	if err != nil {
		return nil, err
	}
	return &SailTrim{
		svc:  lightsail.New(sess),
		conf: c,
	}, nil
}

type DeployOption struct {
	DryRun *bool
}

func (s *SailTrim) Update(ctx context.Context) error {
	sv, err := s.conf.loadService()
	if err != nil {
		return errors.Wrap(err, "failed to load service config")
	}
	if out, err := s.svc.UpdateContainerServiceWithContext(ctx, &lightsail.UpdateContainerServiceInput{
		Power:       sv.Power,
		Scale:       sv.Scale,
		ServiceName: sv.ContainerServiceName,
	}); err != nil {
		return errors.Wrap(err, "failed to update service")
	} else {
		log.Printf("[info] update service: %s", out.String())
	}
	return nil
}

func (s *SailTrim) Deploy(ctx context.Context) error {
	sv, err := s.conf.loadService()
	if err != nil {
		return errors.Wrap(err, "failed to load service config")
	}
	if _, err = s.svc.GetContainerServicesWithContext(ctx, &lightsail.GetContainerServicesInput{
		ServiceName: sv.ContainerServiceName,
	}); err != nil {
		return s.create(sv)
	}

	dp, err := s.conf.loadDeployment()
	if err != nil {
		return errors.Wrap(err, "failed to load deployment config")
	}
	if out, err := s.svc.CreateContainerServiceDeploymentWithContext(ctx, &lightsail.CreateContainerServiceDeploymentInput{
		ServiceName: sv.ContainerServiceName,
		Containers:  dp.Containers,
		PublicEndpoint: &lightsail.EndpointRequest{
			ContainerName: dp.PublicEndpoint.ContainerName,
			ContainerPort: dp.PublicEndpoint.ContainerPort,
			HealthCheck:   dp.PublicEndpoint.HealthCheck,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to create deployment")
	} else {
		log.Printf("[info] create deployment: %s", out.String())
	}
	return nil
}

func (s *SailTrim) create(sv *lightsail.ContainerService) error {
	return errors.New("TODO")
}

func (s *SailTrim) Dump(ctx context.Context, name string) error {
	svOut, err := s.svc.GetContainerServices(&lightsail.GetContainerServicesInput{
		ServiceName: aws.String(name),
	})
	if err != nil {
		return errors.Wrap(err, "failed to get container service")
	}
	if err := s.conf.dumpService(svOut.ContainerServices[0]); err != nil {
		return errors.Wrap(err, "failed to get container service")
	}

	dpOut, err := s.svc.GetContainerServiceDeploymentsWithContext(ctx, &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(name),
	})
	if err != nil {
		return errors.Wrap(err, "failed to get container service deployment")
	}
	if err := s.conf.dumpDeployment(dpOut.Deployments[0]); err != nil {
		return errors.Wrap(err, "failed to dump deployment")
	}
	return nil
}
