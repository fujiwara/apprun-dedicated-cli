package cli

import (
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/apprun-dedicated-api-go/apis/version"
)

// versionDetailToDefinition converts SDK VersionDetail to ApplicationDefinition.
func versionDetailToDefinition(v *version.VersionDetail) *ApplicationDefinition {
	def := &ApplicationDefinition{
		CPU:               v.CPU,
		Memory:            v.Memory,
		ScalingMode:       string(v.ScalingMode),
		FixedScale:        v.FixedScale,
		MinScale:          v.MinScale,
		MaxScale:          v.MaxScale,
		ScaleInThreshold:  v.ScaleInThreshold,
		ScaleOutThreshold: v.ScaleOutThreshold,
		Image:             v.Image,
		Cmd:               v.Cmd,
		RegistryUsername:  v.RegistryUsername,
	}

	for _, ep := range v.ExposedPorts {
		port := ExposedPort{
			TargetPort:     int(ep.TargetPort),
			UseLetsEncrypt: ep.UseLetsEncrypt,
			Host:           ep.Host,
		}
		if ep.LoadBalancerPort != nil {
			lbp := int(*ep.LoadBalancerPort)
			port.LoadBalancerPort = &lbp
		}
		if ep.HealthCheck != nil {
			port.HealthCheck = &HealthCheck{
				Path:            ep.HealthCheck.GetPath(),
				IntervalSeconds: ep.HealthCheck.GetIntervalSeconds(),
				TimeoutSeconds:  ep.HealthCheck.GetTimeoutSeconds(),
			}
		}
		def.ExposedPorts = append(def.ExposedPorts, port)
	}

	for _, ev := range v.EnvVars {
		envVar := EnvVar{
			Key:    ev.Key,
			Secret: ev.Secret,
		}
		if ev.Value != nil {
			envVar.Value = *ev.Value
		}
		def.Env = append(def.Env, envVar)
	}

	return def
}

// definitionToCreateParams converts ApplicationDefinition to SDK CreateParams.
func definitionToCreateParams(def *ApplicationDefinition) version.CreateParams {
	params := version.CreateParams{
		CPU:               def.CPU,
		Memory:            def.Memory,
		ScalingMode:       v1.ScalingMode(def.ScalingMode),
		Image:             def.Image,
		Cmd:               def.Cmd,
		FixedScale:        def.FixedScale,
		MinScale:          def.MinScale,
		MaxScale:          def.MaxScale,
		ScaleInThreshold:  def.ScaleInThreshold,
		ScaleOutThreshold: def.ScaleOutThreshold,
	}

	if def.RegistryUsername != nil {
		params.RegistryUsername = def.RegistryUsername
	}
	if def.RegistryPassword != nil {
		params.RegistryPassword = def.RegistryPassword
	}

	for _, ep := range def.ExposedPorts {
		port := version.ExposedPort{
			TargetPort:     v1.Port(ep.TargetPort),
			UseLetsEncrypt: ep.UseLetsEncrypt,
			Host:           ep.Host,
		}
		if ep.LoadBalancerPort != nil {
			lbPort := v1.Port(*ep.LoadBalancerPort)
			port.LoadBalancerPort = &lbPort
		}
		if ep.HealthCheck != nil {
			hc := v1.HealthCheck{}
			hc.SetPath(ep.HealthCheck.Path)
			hc.SetIntervalSeconds(ep.HealthCheck.IntervalSeconds)
			hc.SetTimeoutSeconds(ep.HealthCheck.TimeoutSeconds)
			port.HealthCheck = &hc
		}
		params.ExposedPorts = append(params.ExposedPorts, port)
	}

	for _, ev := range def.Env {
		envVar := version.EnvironmentVariable{
			Key:    ev.Key,
			Secret: ev.Secret,
		}
		if ev.Value != "" {
			envVar.Value = &ev.Value
		}
		params.EnvVars = append(params.EnvVars, envVar)
	}

	return params
}
