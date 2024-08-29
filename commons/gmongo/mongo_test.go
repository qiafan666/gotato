package gmongo

//func Test_connectWithRetry(t *testing.T) {
//	type args struct {
//		ctx         context.Context
//		clientOpts  *options.ClientOptions
//		maxRetry    int
//		connTimeout time.Duration
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    *gmongo.Client
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := connectWithRetry(tt.args.ctx, tt.args.clientOpts, tt.args.maxRetry, tt.args.connTimeout)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("connectWithRetry() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("connectWithRetry() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestCheckMongo(t *testing.T) {
//	type args struct {
//		ctx    context.Context
//		config *Config
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := CheckMongo(tt.args.ctx, tt.args.config); (err != nil) != tt.wantErr {
//				t.Errorf("CheckMongo() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
