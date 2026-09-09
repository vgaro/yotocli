package yoto

import "encoding/json"

// The API says far more about a card than this client models: the cover art, the
// playback config, a chapter's ambient light, the publisher's copyrights. An
// update replaces the card with the document it is sent, so a field dropped on
// the way in is erased on the way out - which is how changing a track list also
// cleared a card's cover image.
//
// So every type that maps part of a card keeps the fields it does not model and
// writes them back untouched. Each one needs its own pair of methods, since Go
// has no way to share them; they all do the same two things, and the work is in
// capture and merge.

// extras holds the fields of a JSON object that the type reading it does not
// model, exactly as they arrived.
type extras map[string]json.RawMessage

// capture records the fields data holds that known does not model. known must be
// the value data was just unmarshalled into, so that its own fields can be told
// apart from the rest.
func capture(data []byte, known any, extra *extras) error {
	rest := extras{}
	if err := json.Unmarshal(data, &rest); err != nil {
		return err
	}

	out, err := json.Marshal(known)
	if err != nil {
		return err
	}
	var modelled map[string]json.RawMessage
	if err := json.Unmarshal(out, &modelled); err != nil {
		return err
	}
	for field := range modelled {
		delete(rest, field)
	}

	if len(rest) == 0 {
		rest = nil
	}
	*extra = rest
	return nil
}

// merge writes the fields a type models alongside the ones it does not.
func merge(known any, extra extras) ([]byte, error) {
	out, err := json.Marshal(known)
	if err != nil || len(extra) == 0 {
		return out, err
	}

	fields := make(map[string]json.RawMessage, len(extra))
	for field, value := range extra {
		fields[field] = value
	}
	var modelled map[string]json.RawMessage
	if err := json.Unmarshal(out, &modelled); err != nil {
		return nil, err
	}
	for field, value := range modelled {
		fields[field] = value
	}
	return json.Marshal(fields)
}

func (c *Card) UnmarshalJSON(data []byte) error {
	type card Card
	if err := json.Unmarshal(data, (*card)(c)); err != nil {
		return err
	}
	return capture(data, card(*c), &c.extra)
}

func (c Card) MarshalJSON() ([]byte, error) {
	type card Card
	return merge(card(c), c.extra)
}

func (c *Content) UnmarshalJSON(data []byte) error {
	type content Content
	if err := json.Unmarshal(data, (*content)(c)); err != nil {
		return err
	}
	return capture(data, content(*c), &c.extra)
}

func (c Content) MarshalJSON() ([]byte, error) {
	type content Content
	return merge(content(c), c.extra)
}

func (c *Chapter) UnmarshalJSON(data []byte) error {
	type chapter Chapter
	if err := json.Unmarshal(data, (*chapter)(c)); err != nil {
		return err
	}
	return capture(data, chapter(*c), &c.extra)
}

func (c Chapter) MarshalJSON() ([]byte, error) {
	type chapter Chapter
	return merge(chapter(c), c.extra)
}

func (t *Track) UnmarshalJSON(data []byte) error {
	type track Track
	if err := json.Unmarshal(data, (*track)(t)); err != nil {
		return err
	}
	return capture(data, track(*t), &t.extra)
}

func (t Track) MarshalJSON() ([]byte, error) {
	type track Track
	return merge(track(t), t.extra)
}

func (d *Display) UnmarshalJSON(data []byte) error {
	type display Display
	if err := json.Unmarshal(data, (*display)(d)); err != nil {
		return err
	}
	return capture(data, display(*d), &d.extra)
}

func (d Display) MarshalJSON() ([]byte, error) {
	// A card that has never had icons has "display": null on its tracks, and
	// writing an empty object in its place would be a change where none was
	// asked for. Display is a value rather than a pointer, so this is where the
	// distinction has to be kept.
	if d.Icon16x16 == "" && len(d.extra) == 0 {
		return []byte("null"), nil
	}

	type display Display
	return merge(display(d), d.extra)
}

func (m *Metadata) UnmarshalJSON(data []byte) error {
	type metadata Metadata
	if err := json.Unmarshal(data, (*metadata)(m)); err != nil {
		return err
	}
	return capture(data, metadata(*m), &m.extra)
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	type metadata Metadata
	return merge(metadata(m), m.extra)
}

func (m *Media) UnmarshalJSON(data []byte) error {
	type media Media
	if err := json.Unmarshal(data, (*media)(m)); err != nil {
		return err
	}
	return capture(data, media(*m), &m.extra)
}

func (m Media) MarshalJSON() ([]byte, error) {
	type media Media
	return merge(media(m), m.extra)
}

// InheritFrom returns c presented the way prev was.
//
// prev is the chapter already on the card for the same audio, so it may carry
// work done in the Yoto app that no source file knows about: the icon above all,
// but also whatever else the app records there and this client does not model.
// What describes the audio itself - title, duration, size, URL - is left as c has
// it, since that is what the source now says.
func (c Chapter) InheritFrom(prev Chapter) Chapter {
	c.Display = prev.Display
	c.extra = prev.extra

	// Copied rather than written through: c's tracks may be shared with the
	// chapter a caller built them from.
	tracks := make([]Track, len(c.Tracks))
	copy(tracks, c.Tracks)
	for i := range tracks {
		if i >= len(prev.Tracks) {
			break
		}
		tracks[i].Display = prev.Tracks[i].Display
		tracks[i].extra = prev.Tracks[i].extra
	}
	c.Tracks = tracks
	return c
}
