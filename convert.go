package cli

import (
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/apprun-dedicated-api-go/apis/version"
)

// versionDetailToDefinition converts SDK VersionDetail to our ApplicationDefinition
func versionDetailToDefinition(v *version.VersionDetail) *ApplicationDefinition {
	def := &ApplicationDefinition{
		CPU:         int(v.CPU),
		Memory:      int(v.Memory),
		ScalingMode: string(v.ScalingMode),
		Image:       imageFromString(v.Image),
		Cmd:         v.Cmd,
	}

	if v.RegistryUsername != nil {
		def.Image.RegistryUsername = *v.RegistryUsername
	}

	if v.FixedScale != nil {
		val := int(*v.FixedScale)
		def.FixedScale = &val
	}
	if v.MinScale != nil {
		val := int(*v.MinScale)
		def.MinScale = &val
	}
	if v.MaxScale != nil {
		val := int(*v.MaxScale)
		def.MaxScale = &val
	}
	if v.ScaleInThreshold != nil {
		val := int(*v.ScaleInThreshold)
		def.ScaleInThreshold = &val
	}
	if v.ScaleOutThreshold != nil {
		val := int(*v.ScaleOutThreshold)
		def.ScaleOutThreshold = &val
	}

	for _, ep := range v.ExposedPorts {
		port := ExposedPort{
			TargetPort:     int(ep.TargetPort),
			UseLetsEncrypt: ep.UseLetsEncrypt,
			Host:           ep.Host,
		}
		if ep.LoadBalancerPort != nil {
			port.LoadBalancerPort = int(*ep.LoadBalancerPort)
		}
		if ep.HealthCheck != nil {
			port.HealthCheck = &HealthCheck{
				Path:            ep.HealthCheck.GetPath(),
				IntervalSeconds: int(ep.HealthCheck.GetIntervalSeconds()),
				TimeoutSeconds:  int(ep.HealthCheck.GetTimeoutSeconds()),
			}
		}
		def.ExposedPorts = append(def.ExposedPorts, port)
	}

	for _, ev := range v.EnvVars {
		envVar := EnvironmentVariable{
			Key:    ev.Key,
			Secret: ev.Secret,
		}
		if ev.Value != nil {
			envVar.Value = *ev.Value
		}
		def.EnvironmentVariables = append(def.EnvironmentVariables, envVar)
	}

	return def
}

// definitionToCreateParams converts our ApplicationDefinition to SDK CreateParams
func definitionToCreateParams(def *ApplicationDefinition) version.CreateParams {
	params := version.CreateParams{
		CPU:         int64(def.CPU),
		Memory:      int64(def.Memory),
		ScalingMode: v1.ScalingMode(def.ScalingMode),
		Image:       def.Image.Path + ":" + def.Image.Tag,
		Cmd:         def.Cmd,
	}

	if def.Image.RegistryUsername != "" {
		params.RegistryUsername = &def.Image.RegistryUsername
	}
	if def.Image.RegistryPassword != "" {
		params.RegistryPassword = &def.Image.RegistryPassword
		params.RegistryPasswordAction = v1.RegistryPasswordActionNew
	}
	// RegistryPasswordAction is set by caller (Keep for updates, omit for first version)

	if def.FixedScale != nil {
		val := int32(*def.FixedScale)
		params.FixedScale = &val
	}
	if def.MinScale != nil {
		val := int32(*def.MinScale)
		params.MinScale = &val
	}
	if def.MaxScale != nil {
		val := int32(*def.MaxScale)
		params.MaxScale = &val
	}
	if def.ScaleInThreshold != nil {
		val := int32(*def.ScaleInThreshold)
		params.ScaleInThreshold = &val
	}
	if def.ScaleOutThreshold != nil {
		val := int32(*def.ScaleOutThreshold)
		params.ScaleOutThreshold = &val
	}

	for _, ep := range def.ExposedPorts {
		port := version.ExposedPort{
			TargetPort:     v1.Port(ep.TargetPort),
			UseLetsEncrypt: ep.UseLetsEncrypt,
			Host:           ep.Host,
		}
		if ep.LoadBalancerPort != 0 {
			lbPort := v1.Port(ep.LoadBalancerPort)
			port.LoadBalancerPort = &lbPort
		}
		if ep.HealthCheck != nil {
			hc := v1.HealthCheck{}
			hc.SetPath(ep.HealthCheck.Path)
			hc.SetIntervalSeconds(int32(ep.HealthCheck.IntervalSeconds))
			hc.SetTimeoutSeconds(int32(ep.HealthCheck.TimeoutSeconds))
			port.HealthCheck = &hc
		}
		params.ExposedPorts = append(params.ExposedPorts, port)
	}

	for _, ev := range def.EnvironmentVariables {
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

// imageFromString parses "path:tag" into ImageDefinition
func imageFromString(image string) ImageDefinition {
	for i := len(image) - 1; i >= 0; i-- {
		if image[i] == ':' {
			return ImageDefinition{
				Path: image[:i],
				Tag:  image[i+1:],
			}
		}
		if image[i] == '/' {
			break
		}
	}
	return ImageDefinition{Path: image, Tag: "latest"}
}
