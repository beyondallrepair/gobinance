package gobinance

import (
	"fmt"
	"reflect"
	"testing"
)

type validatable interface {
	Validate() error
}

func testValidatableEnum(t *testing.T, invalidValue validatable, validValues ...validatable) {
	typInv := reflect.TypeOf(invalidValue)
	test := func(v validatable, shouldError bool) {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			t.Parallel()
			if typV := reflect.TypeOf(v); typV != typInv {
				t.Fatalf("incorrect type passed for value %v. expected %v but got %v", v, typInv, typV)
			}
			err := v.Validate()
			if shouldError {
				if err == nil {
					t.Errorf("expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	test(invalidValue, true)
	for _, v := range validValues {
		v := v
		test(v, false)
	}
}

func TestOrderStatus_Validate(t *testing.T) {
	testValidatableEnum(t,
		OrderStatus("invalid"),
		OrderStatusNew,
		OrderStatusPartiallyFilled,
		OrderStatusFilled,
		OrderStatusCanceled,
		OrderStatusRejected,
		OrderStatusExpired,
	)
}

func TestOrderResponseType_Validate(t *testing.T) {
	testValidatableEnum(t,
		OrderResponseType("invalid"),
		OrderResponseTypeFull,
		OrderResponseTypeAck,
		OrderResponseTypeResult,
	)
}

func TestOrderType_Validate(t *testing.T) {
	testValidatableEnum(t,
		OrderType("invalid"),
		OrderTypeLimit,
		OrderTypeMarket,
		OrderTypeStopLoss,
		OrderTypeStopLossLimit,
		OrderTypeTakeProfit,
		OrderTypeTakeProfitLimit,
		OrderTypeLimitMaker,
	)
}

func TestOrderSide_Validate(t *testing.T) {
	testValidatableEnum(t,
		OrderSide("invalid"),
		OrderSideSell,
		OrderSideBuy,
	)
}

func TestTimeInForce_Validate(t *testing.T) {
	testValidatableEnum(t,
		TimeInForce("invalid"),
		TimeInForceGoodTilCanceled,
		TimeInForceImmediateOrCancel,
		TimeInForceFillOrKill,
	)
}