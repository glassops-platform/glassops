package salesforce

import (
	"errors"
	"os"
	"runtime"
	"sync"

	"github.com/glassops-platform/glassops/internal/providers"
)

var isWindows = runtime.GOOS == "windows"

// sessionAlias is the sf alias of the target inside a session's isolated state.
const sessionAlias = "glassops-target"

// session is one login. Its home dir holds the key file and sf's state for this login only.
type session struct {
	target providers.Target
	// home is the private temp dir. sf sees it as the user's home.
	home string
	// alias is the --target-org value for every sf call in this session.
	alias string
	// stop detaches the context cleanup that Authenticate registered. It may be nil.
	stop func() bool

	closeOnce sync.Once
	closeErr  error
}

func (s *session) Target() providers.Target { return s.target }

// Close removes the session dir. It is safe to call more than once.
func (s *session) Close() error {
	s.closeOnce.Do(func() {
		if s.stop != nil {
			s.stop()
		}
		if s.home != "" {
			s.closeErr = os.RemoveAll(s.home)
		}
	})
	return s.closeErr
}

// env points sf at the session's home, so its logins stay out of the user's real store.
// Plugins are unaffected: sf finds them through its data dir (SF_DATA_DIR in the image,
// %LOCALAPPDATA%\sf on Windows).
func (s *session) env() []string {
	return []string{"HOME=" + s.home, "USERPROFILE=" + s.home}
}

// asSession unwraps a providers.Session made by this provider.
func asSession(s providers.Session) (*session, bool) {
	ss, ok := s.(*session)
	return ss, ok
}

// SessionEnv points sf in an exec step's command at this session: the session's home holds
// the login, and SF_TARGET_ORG makes its alias the default org, so a plain "sf data query"
// runs against the target. The token itself is never put in the environment.
func (p *Provider) SessionEnv(s providers.Session) ([]string, error) {
	ss, ok := asSession(s)
	if !ok || ss.home == "" {
		return nil, errors.New("not a Salesforce session")
	}
	return append(ss.env(), "SF_TARGET_ORG="+ss.alias), nil
}
