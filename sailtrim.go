package sailtrim

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/Songmu/prompter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/pkg/errors"
)

// SailTrim represents an application.
type SailTrim struct {
	svc  *lightsail.Lightsail
	conf *Config
}

// New creates a SailTrim instance.
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

// DeployOption represents options for Deploy.
type DeployOption struct {
	DryRun *bool
}

// Update updates container service attributes.
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

// Deploy deploies a container service with new deployment.
func (s *SailTrim) Deploy(ctx context.Context) error {
	sv, err := s.conf.loadService()
	if err != nil {
		return errors.Wrap(err, "failed to load service config")
	}
	if _, err = s.svc.GetContainerServicesWithContext(ctx, &lightsail.GetContainerServicesInput{
		ServiceName: sv.ContainerServiceName,
	}); err != nil {
		return s.create(ctx, *sv.ContainerServiceName)
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

// Init initializes a container service.
func (s *SailTrim) Init(ctx context.Context, serviceName string) error {
	svOut, err := s.svc.GetContainerServices(&lightsail.GetContainerServicesInput{
		ServiceName: aws.String(serviceName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case lightsail.ErrCodeNotFoundException:
				ok := prompter.YN(fmt.Sprintf("%s is not exist. Create new configuration files?", serviceName), false)
				if ok {
					return s.initConfigurations(ctx, serviceName)
				}
			}
		}
		return errors.Wrap(err, "failed to get container service")
	}
	if err := s.conf.dumpService(svOut.ContainerServices[0]); err != nil {
		return errors.Wrap(err, "failed to get container service")
	}

	dpOut, err := s.svc.GetContainerServiceDeploymentsWithContext(ctx, &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(serviceName),
	})
	if err != nil {
		return errors.Wrap(err, "failed to get container service deployment")
	}
	if err := s.conf.dumpDeployment(dpOut.Deployments[0]); err != nil {
		return errors.Wrap(err, "failed to dump deployment")
	}
	return nil
}

var (
	scaleChoices = []string{"1", "2", "3", "4"}
	scaleDefault = "1"
	powerChoices = []string{"nano", "micro", "small", "medium", "large", "xlarge"}
	powerDefault = "micro"
)

func (s *SailTrim) initConfigurations(ctx context.Context, serviceName string) error {
	sv := lightsail.ContainerService{
		ContainerServiceName: aws.String(serviceName),
	}
	sv.Power = aws.String(
		prompter.Choose("Choose the power", powerChoices, powerDefault),
	)
	if sc := prompter.Choose("Choose the scale", scaleChoices, scaleDefault); sc != "" {
		if scale, err := strconv.ParseInt(sc, 10, 64); err != nil {
			return fmt.Errorf("invalid scale %s", sc)
		} else {
			sv.Scale = aws.Int64(scale)
		}
	}

	dp := lightsail.ContainerServiceDeployment{
		Containers: make(map[string]*lightsail.Container, 0),
	}
	endpoints := []string{"No endpoint"}
	for {
		containerName := prompter.Prompt("Container name", "")
		c := lightsail.Container{
			Environment: make(map[string]*string, 0),
			Ports:       make(map[string]*string, 0),
		}
		c.Image = aws.String(
			prompter.Prompt("Image", ""),
		)
		if cmd := prompter.Prompt("Launch command", ""); cmd != "" {
			c.Command = aws.StringSlice([]string{cmd})
		}
		for prompter.YN("Add an environment variable?", false) {
			c.Environment[prompter.Prompt("Key", "")] = aws.String(
				prompter.Prompt("Value", ""),
			)
		}
		var endpointsAdded bool
		for prompter.YN("Add an open port?", false) {
			port := prompter.Prompt("Port", "")
			proto := prompter.Choose("Protocol", []string{"HTTP", "HTTPS", "TCP", "UDP"}, "HTTP")
			c.Ports[port] = aws.String(proto)
			if (proto == "HTTP" || proto == "HTTPS") && !endpointsAdded {
				endpoints = append(endpoints, containerName)
				endpointsAdded = true
			}
		}
		if prompter.YN("Add container entry?", false) {
			continue
		} else {
			dp.Containers[containerName] = &c
			break
		}
	}
	endpoint := prompter.Choose("Public endpoint container", endpoints, endpoints[0])
	if endpoint != endpoints[0] {
		ports := []string{}
		for p := range dp.Containers[endpoint].Ports {
			port := p
			ports = append(ports, port)
		}
		sort.SliceStable(ports, func(i, j int) bool {
			in, err := strconv.ParseInt(ports[i], 10, 64)
			if err != nil {
				return false
			}
			jn, err := strconv.ParseInt(ports[j], 10, 64)
			if err != nil {
				return false
			}
			return in < jn
		})
		port := prompter.Choose("Public endpoint port", ports, ports[0])
		pn, _ := strconv.ParseInt(port, 10, 64)
		dp.PublicEndpoint = &lightsail.ContainerServiceEndpoint{
			ContainerName: aws.String(endpoint),
			ContainerPort: aws.Int64(pn),
			HealthCheck: &lightsail.ContainerServiceHealthCheckConfig{
				HealthyThreshold:   aws.Int64(2),
				IntervalSeconds:    aws.Int64(5),
				Path:               aws.String("/"),
				SuccessCodes:       aws.String("200-499"),
				TimeoutSeconds:     aws.Int64(2),
				UnhealthyThreshold: aws.Int64(2),
			},
		}
	}

	if err := s.conf.dumpService(&sv); err != nil {
		return err
	}
	if err := s.conf.dumpDeployment(&dp); err != nil {
		return err
	}
	return nil
}

func (s *SailTrim) create(ctx context.Context, serviceName string) error {
	log.Printf("[info] service and deployment will be created as below.")
	if err := s.conf.printService(os.Stdout); err != nil {
		return err
	}
	if err := s.conf.printDeployment(os.Stdout); err != nil {
		return err
	}

	sv, err := s.conf.loadService()
	if err != nil {
		return err
	}
	dp, err := s.conf.loadDeployment()
	if err != nil {
		return err
	}

	if prompter.YN("Do you create container service?", false) {
		log.Println("[info] creating container service...")
		_, err := s.svc.CreateContainerServiceWithContext(
			ctx,
			&lightsail.CreateContainerServiceInput{
				ServiceName: sv.ContainerServiceName,
				Power:       sv.Power,
				Scale:       sv.Scale,
				Deployment: &lightsail.ContainerServiceDeploymentRequest{
					Containers: dp.Containers,
					PublicEndpoint: &lightsail.EndpointRequest{
						ContainerName: dp.PublicEndpoint.ContainerName,
						ContainerPort: dp.PublicEndpoint.ContainerPort,
						HealthCheck:   dp.PublicEndpoint.HealthCheck,
					},
				},
			},
		)
		return err
	}
	return nil
}
