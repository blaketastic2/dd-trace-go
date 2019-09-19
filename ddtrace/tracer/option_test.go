// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

package tracer

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/globalconfig"
)

func withTransport(t transport) StartOption {
	return func(c *config) {
		c.transport = t
	}
}

func TestTracerOptionsDefaults(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		assert := assert.New(t)
		var c config
		defaults(&c)
		assert.Equal(float64(1), c.sampler.(RateSampler).Rate())
		assert.Equal("tracer.test", c.serviceName)
		assert.Equal("localhost:8126", c.agentAddr)
		assert.Equal(nil, c.httpRoundTripper)
	})

	t.Run("analytics", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(math.IsNaN(globalconfig.AnalyticsRate()))
		newTracer(WithAnalyticsRate(0.5))
		assert.Equal(0.5, globalconfig.AnalyticsRate())
		newTracer(WithAnalytics(false))
		assert.True(math.IsNaN(globalconfig.AnalyticsRate()))
		newTracer(WithAnalytics(true))
		assert.Equal(1., globalconfig.AnalyticsRate())
	})

	t.Run("other", func(t *testing.T) {
		assert := assert.New(t)
		tracer := newTracer(
			WithSampler(NewRateSampler(0.5)),
			WithServiceName("api-intake"),
			WithAgentAddr("ddagent.consul.local:58126"),
			WithGlobalTag("k", "v"),
			WithDebugMode(true),
		)
		c := tracer.config
		assert.Equal(float64(0.5), c.sampler.(RateSampler).Rate())
		assert.Equal("api-intake", c.serviceName)
		assert.Equal("ddagent.consul.local:58126", c.agentAddr)
		assert.NotNil(c.globalTags)
		assert.Equal("v", c.globalTags["k"])
		assert.True(c.debug)
	})

	t.Run("envs", func(t *testing.T) {
		if err := os.Setenv("DD_SERVICE_NAME", "TEST_SERVICE"); err != nil {
			panic("could not set environment variable DD_SERVICE_NAME during testing")
		}
		if err := os.Setenv("DD_TRACE_GLOBAL_TAGS", "env:test, aKey:aVal,bKey:bVal"); err != nil {
			panic("could not set environment variable DD_TRACE_GLOBAL_TAGS during testing")
		}

		assert := assert.New(t)
		var c config
		defaults(&c)
		assert.Equal("TEST_SERVICE", c.serviceName)

		env, ok := c.globalTags["env"]
		assert.True(ok, "has the env key")
		assert.Equal("test", env)

		aVal, ok := c.globalTags["aKey"]
		assert.True(ok, "has aKey key")
		assert.Equal("aVal", aVal)

		bVal, ok := c.globalTags["bKey"]
		assert.True(ok, "has bKey key")
		assert.Equal("bVal", bVal)

		cVal, ok := c.globalTags["cKey"]
		assert.False(ok, "does not have cKey")
		assert.Equal("", cVal)
	})
}
