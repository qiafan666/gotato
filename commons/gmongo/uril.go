package gmongo

import (
	"context"
	"errors"
	"github.com/qiafan666/gotato/commons/gerr"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func IsDBNotFound(err error) bool {
	return errors.Is(mongo.ErrNoDocuments, gerr.Unwrap(err))
}

func basic[T any]() bool {
	var t T
	switch any(t).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, []byte:
		return true
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *[]byte:
		return true
	default:
		return false
	}
}

func anes[T any](ts []T) []any {
	val := make([]any, len(ts))
	for i := range ts {
		val[i] = ts[i]
	}
	return val
}

func findOptionToCountOption(opts []*options.FindOptions) *options.CountOptions {
	return options.Count()
}

func InsertMany[T any](ctx context.Context, coll *mongo.Collection, val []T, opts ...*options.InsertManyOptions) error {
	_, err := coll.InsertMany(ctx, anes(val), opts...)
	if err != nil {
		return gerr.WrapMsg(err, "gmongo insert many")
	}
	return nil
}

func UpdateOne(ctx context.Context, coll *mongo.Collection, filter any, update any, notMatchedErr bool, opts ...*options.UpdateOptions) error {
	res, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return gerr.WrapMsg(err, "gmongo update one")
	}
	if notMatchedErr && res.MatchedCount == 0 {
		return gerr.WrapMsg(mongo.ErrNoDocuments, "gmongo update not matched")
	}
	return nil
}

func UpdateOneResult(ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	res, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo update one")
	}
	return res, nil
}

func UpdateMany(ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	res, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo update many")
	}
	return res, nil
}

func Find[T any](ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.FindOptions) ([]T, error) {
	cur, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo find")
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err = cur.Close(ctx)
		if err != nil {

		}
	}(cur, ctx)
	return Decodes[T](ctx, cur)
}

func FindOne[T any](ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.FindOneOptions) (res T, err error) {
	cur := coll.FindOne(ctx, filter, opts...)
	if err = cur.Err(); err != nil {
		return res, gerr.WrapMsg(err, "gmongo find one")
	}
	return DecodeOne[T](cur.Decode)
}

func FindOneAndUpdate[T any](ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...*options.FindOneAndUpdateOptions) (res T, err error) {
	result := coll.FindOneAndUpdate(ctx, filter, update, opts...)
	if err = result.Err(); err != nil {
		return res, gerr.WrapMsg(err, "gmongo find one and update")
	}
	return DecodeOne[T](result.Decode)
}

func FindPage[T any](ctx context.Context, coll *mongo.Collection, filter any, pageNum, pageSize int, opts ...*options.FindOptions) (int64, []T, error) {
	count, err := Count(ctx, coll, filter, findOptionToCountOption(opts))
	if err != nil {
		return 0, nil, gerr.WrapMsg(err, "gmongo failed to count documents in collection")
	}
	if count == 0 {
		return count, nil, nil
	}
	skip := int64(pageNum-1) * int64(pageSize)
	if skip < 0 || skip >= count || pageSize <= 0 {
		return count, nil, nil
	}
	opt := options.Find().SetSkip(skip).SetLimit(int64(pageSize))
	res, err := Find[T](ctx, coll, filter, append(opts, opt)...)
	if err != nil {
		return 0, nil, err
	}
	return count, res, nil
}

func FindPageOnly[T any](ctx context.Context, coll *mongo.Collection, filter any, pageNum, pageSize int, opts ...*options.FindOptions) ([]T, error) {
	skip := int64(pageNum-1) * int64(pageSize)
	if skip < 0 || pageSize <= 0 {
		return nil, nil
	}
	opt := options.Find().SetSkip(skip).SetLimit(int64(pageSize))
	return Find[T](ctx, coll, filter, append(opts, opt)...)
}

func Count(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.CountOptions) (int64, error) {
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, gerr.WrapMsg(err, "gmongo count")
	}
	return count, nil
}

func Exist(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.CountOptions) (bool, error) {
	opts = append(opts, options.Count().SetLimit(1))
	count, err := Count(ctx, coll, filter, opts...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteOne(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.DeleteOptions) error {
	if _, err := coll.DeleteOne(ctx, filter, opts...); err != nil {
		return gerr.WrapMsg(err, "gmongo delete one")
	}
	return nil
}

func DeleteOneResult(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	res, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo delete one")
	}
	return res, nil
}

func DeleteMany(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.DeleteOptions) error {
	if _, err := coll.DeleteMany(ctx, filter, opts...); err != nil {
		return gerr.WrapMsg(err, "gmongo delete many")
	}
	return nil
}

func DeleteManyResult(ctx context.Context, coll *mongo.Collection, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	res, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo delete many")
	}
	return res, nil
}

func Aggregate[T any](ctx context.Context, coll *mongo.Collection, pipeline any, opts ...*options.AggregateOptions) ([]T, error) {
	cur, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, gerr.WrapMsg(err, "gmongo aggregate")
	}
	defer cur.Close(ctx)
	return Decodes[T](ctx, cur)
}

func Decodes[T any](ctx context.Context, cur *mongo.Cursor) ([]T, error) {
	var res []T
	if basic[T]() {
		var temp []map[string]T
		if err := cur.All(ctx, &temp); err != nil {
			return nil, gerr.WrapMsg(err, "gmongo decodes")
		}
		res = make([]T, 0, len(temp))
		for _, m := range temp {
			if len(m) != 1 {
				return nil, gerr.New("gmongo find result len(m) != 1")
			}
			for _, t := range m {
				res = append(res, t)
			}
		}
	} else {
		if err := cur.All(ctx, &res); err != nil {
			return nil, gerr.WrapMsg(err, "gmongo all")
		}
	}
	return res, nil
}

func DecodeOne[T any](decoder func(v any) error) (res T, err error) {
	if basic[T]() {
		var temp map[string]T
		if err = decoder(&temp); err != nil {
			err = gerr.WrapMsg(err, "gmongo decodes one")
			return
		}
		if len(temp) != 1 {
			err = gerr.New("gmongo find result len(m) != 1")
			return
		}
		for k := range temp {
			res = temp[k]
		}
	} else {
		if err = decoder(&res); err != nil {
			err = gerr.WrapMsg(err, "gmongo decoder")
			return
		}
	}
	return
}

func Ignore[T any](_ T, err error) error {
	return err
}

func IgnoreWarp[T any](_ T, err error) error {
	if err != nil {
		return gerr.WrapMsg(err, "gmongo ignore")
	}
	return err
}

func IncrVersion(dbs ...func() error) error {
	for _, fn := range dbs {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
