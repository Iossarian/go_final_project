package bucket

import (
	testing "testing"

	config "abf/internal/config"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	cfg := config.NewConfig("./config.test.yml")

	t.Run("tests login bruteforce", func(t *testing.T) {
		s := NewStorage(cfg)
		for i := 0; i < 2*s.loginMaxLoad; i++ {
			err := s.Check("testLogin", "", "")
			if i < s.loginMaxLoad {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrRejected)
			}
		}
	})

	t.Run("tests password bruteforce", func(t *testing.T) {
		s := NewStorage(cfg)
		for i := 0; i < 2*s.passwordMaxLoad; i++ {
			err := s.Check("", "testPassword", "")
			if i < s.passwordMaxLoad {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrRejected)
			}
		}
	})

	t.Run("tests login bruteforce", func(t *testing.T) {
		s := NewStorage(cfg)
		for i := 0; i < 2*s.ipMaxLoad; i++ {
			err := s.Check("", "", "127.0.0.1")
			if i < s.ipMaxLoad {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrRejected)
			}
		}
	})
}

func TestStorageParallel(t *testing.T) {
	t.Parallel()
	t.Run("tests parallel", func(t *testing.T) {
		t.Parallel()
		cfg := config.NewConfig("./config.parallel.yml")
		s := NewStorage(cfg)
		errCount := 0
		for i := 0; i < 100; i++ {
			if err := s.Check("testLogin", "", ""); err != nil {
				errCount++
			}
		}

		require.Equal(t, errCount, 40)
	})
}
