package sshclient

import (
	"context"

	"github.com/google/uuid"
)

func (sc *sshClient) NewFriendReq(ctx context.Context, nickname string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(nickname)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("new-friend", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) AcceptFriendReq(ctx context.Context, iden Identity) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(iden)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("accept-friend", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) DenyFriendReq(ctx context.Context, id uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(id)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("deny-friend", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) DeleteFromFriends(ctx context.Context, id uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(id)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("delete-friend", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) BlockUser(ctx context.Context, nickname string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		data, err := MarshData(nickname)
		if err != nil {
			return nil, err
		}
		status, payload, err := sc.client.SendRequest("block-user", true, data)
		if err != nil {
			return nil, err
		}
		if !status {
			return nil, castErr(payload)
		}

		return payload, nil
	}
}
func (sc *sshClient) UnblockUser(ctx context.Context, nickname string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		data, err := MarshData(nickname)
		if err != nil {
			return nil, err
		}
		status, payload, err := sc.client.SendRequest("unblock-user", true, data)
		if err != nil {
			return nil, err
		}
		if !status {
			return nil, castErr(payload)
		}
		return payload, nil
	}
}

func (sc *sshClient) UpdateCurrentConnects(ctx context.Context, conns []Identity) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(conns)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("conns-update", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) SetTagline(ctx context.Context, tagline string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(tagline)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("set-tagline", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) NewNickname(ctx context.Context, nickname string, password []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var d struct {
			NewNickname string `json:"new-nickname"`
			Password    []byte `json:"password"`
		}
		d.NewNickname = nickname
		d.Password = password
		data, err := MarshData(d)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("new-nickname", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) NewPassword(ctx context.Context, oldPassword, newPassword []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var pp struct {
			OldPassword []byte `json:"old-password"`
			NewPassword []byte `json:"new-password"`
		}
		pp.OldPassword = oldPassword
		pp.NewPassword = newPassword
		data, err := MarshData(pp)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("new-password", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) SetColor(ctx context.Context, color string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data, err := MarshData(color)
		if err != nil {
			return err
		}
		status, payload, err := sc.client.SendRequest("set-color", true, data)
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}
