package chat

import "sort"

type ChannelSync struct {
	channels map[string]struct{}
}

func NewChannelSync() ChannelSync {
	return ChannelSync{
		channels: make(map[string]struct{}),
	}
}

func (s *ChannelSync) SetOwnedChannels(channels []string) {
	s.channels = make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		if channel == "" {
			continue
		}
		s.channels[channel] = struct{}{}
	}
}

func (s *ChannelSync) OwnedChannels() []string {
	channels := make([]string, 0, len(s.channels))
	for channel := range s.channels {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return channels
}
