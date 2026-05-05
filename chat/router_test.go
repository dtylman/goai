package chat_test

import (
	"testing"

	"github.com/dtylman/goai/chat"
)

func TestSingleClient_AlwaysResolvesToSame(t *testing.T) {
	mock := &mockClient{response: "ok"}
	router := chat.SingleClient(mock)

	c1, err := router.Resolve("translate")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := router.Resolve("proofread")
	if err != nil {
		t.Fatal(err)
	}
	if c1 != c2 {
		t.Error("expected same client for all roles")
	}
}

func TestMap_ResolvesCorrectClient(t *testing.T) {
	translate := &mockClient{response: "translated"}
	proofread := &mockClient{response: "proofread"}

	router := chat.Map(map[string]chat.Client{
		"translate": translate,
		"proofread": proofread,
	}, nil)

	c, err := router.Resolve("translate")
	if err != nil {
		t.Fatal(err)
	}
	if c != translate {
		t.Error("expected translate client")
	}

	c, err = router.Resolve("proofread")
	if err != nil {
		t.Fatal(err)
	}
	if c != proofread {
		t.Error("expected proofread client")
	}
}

func TestMap_FallsBackToDefault(t *testing.T) {
	fallback := &mockClient{response: "default"}
	router := chat.Map(map[string]chat.Client{}, fallback)

	c, err := router.Resolve("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if c != fallback {
		t.Error("expected fallback client")
	}
}

func TestMap_ErrorsWithoutDefault(t *testing.T) {
	router := chat.Map(map[string]chat.Client{}, nil)

	_, err := router.Resolve("missing")
	if err == nil {
		t.Fatal("expected error for missing role without default")
	}
}
