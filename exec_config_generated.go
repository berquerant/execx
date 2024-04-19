// Code generated by "goconfig -field StdoutConsumer func(Token)|StderrConsumer func(Token)|SplitFunc SplitFunc|SplitSeparator []byte|StdoutWriter io.Writer|StderrWriter io.Writer -option -output exec_config_generated.go -configOption Option"; DO NOT EDIT.

package execx

import "io"

type ConfigItem[T any] struct {
	modified     bool
	value        T
	defaultValue T
}

func (s *ConfigItem[T]) Set(value T) {
	s.modified = true
	s.value = value
}
func (s *ConfigItem[T]) Get() T {
	if s.modified {
		return s.value
	}
	return s.defaultValue
}
func (s *ConfigItem[T]) Default() T {
	return s.defaultValue
}
func (s *ConfigItem[T]) IsModified() bool {
	return s.modified
}
func NewConfigItem[T any](defaultValue T) *ConfigItem[T] {
	return &ConfigItem[T]{
		defaultValue: defaultValue,
	}
}

type Config struct {
	StdoutConsumer *ConfigItem[func(Token)]
	StderrConsumer *ConfigItem[func(Token)]
	SplitFunc      *ConfigItem[SplitFunc]
	SplitSeparator *ConfigItem[[]byte]
	StdoutWriter   *ConfigItem[io.Writer]
	StderrWriter   *ConfigItem[io.Writer]
}
type ConfigBuilder struct {
	stdoutConsumer func(Token)
	stderrConsumer func(Token)
	splitFunc      SplitFunc
	splitSeparator []byte
	stdoutWriter   io.Writer
	stderrWriter   io.Writer
}

func (s *ConfigBuilder) StdoutConsumer(v func(Token)) *ConfigBuilder {
	s.stdoutConsumer = v
	return s
}
func (s *ConfigBuilder) StderrConsumer(v func(Token)) *ConfigBuilder {
	s.stderrConsumer = v
	return s
}
func (s *ConfigBuilder) SplitFunc(v SplitFunc) *ConfigBuilder {
	s.splitFunc = v
	return s
}
func (s *ConfigBuilder) SplitSeparator(v []byte) *ConfigBuilder {
	s.splitSeparator = v
	return s
}
func (s *ConfigBuilder) StdoutWriter(v io.Writer) *ConfigBuilder {
	s.stdoutWriter = v
	return s
}
func (s *ConfigBuilder) StderrWriter(v io.Writer) *ConfigBuilder {
	s.stderrWriter = v
	return s
}
func (s *ConfigBuilder) Build() *Config {
	return &Config{
		StdoutConsumer: NewConfigItem(s.stdoutConsumer),
		StderrConsumer: NewConfigItem(s.stderrConsumer),
		SplitFunc:      NewConfigItem(s.splitFunc),
		SplitSeparator: NewConfigItem(s.splitSeparator),
		StdoutWriter:   NewConfigItem(s.stdoutWriter),
		StderrWriter:   NewConfigItem(s.stderrWriter),
	}
}

func NewConfigBuilder() *ConfigBuilder { return &ConfigBuilder{} }
func (s *Config) Apply(opt ...Option) {
	for _, x := range opt {
		x(s)
	}
}

type Option func(*Config)

func WithStdoutConsumer(v func(Token)) Option {
	return func(c *Config) {
		c.StdoutConsumer.Set(v)
	}
}
func WithStderrConsumer(v func(Token)) Option {
	return func(c *Config) {
		c.StderrConsumer.Set(v)
	}
}
func WithSplitFunc(v SplitFunc) Option {
	return func(c *Config) {
		c.SplitFunc.Set(v)
	}
}
func WithSplitSeparator(v []byte) Option {
	return func(c *Config) {
		c.SplitSeparator.Set(v)
	}
}
func WithStdoutWriter(v io.Writer) Option {
	return func(c *Config) {
		c.StdoutWriter.Set(v)
	}
}
func WithStderrWriter(v io.Writer) Option {
	return func(c *Config) {
		c.StderrWriter.Set(v)
	}
}
