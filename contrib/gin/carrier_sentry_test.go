package tracinggin

import "testing"

func Test_convertSentryTraceToParent(t *testing.T) {
	type args struct {
		sentryTrace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 0",
			args: args{
				sentryTrace: "994f3d6cebfc1b8f19c52a8c687ab5f3-0c77d121fee1c02a-1",
			},
			want: "00-994f3d6cebfc1b8f19c52a8c687ab5f3-0c77d121fee1c02a-01",
		},
		{
			name: "case 1",
			args: args{
				sentryTrace: "00-994f3d6cebfc1b8f19c52a8c687ab5f3-0c77d121fee1c02a-01",
			},
			want: "00-994f3d6cebfc1b8f19c52a8c687ab5f3-0c77d121fee1c02a-01",
		},
		{
			name: "case 2",
			args: args{
				sentryTrace: "",
			},
			want: "",
		},
		{
			name: "case 3",
			args: args{
				sentryTrace: "00-994f3d6cebfc1b8f19c52a8c687ab5f3",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := translateSentryToOpenTelemetry(tt.args.sentryTrace); got != tt.want {
				t.Errorf("translateSentryToOpenTelemetry() = %v, want %v", got, tt.want)
			}
		})
	}
}
