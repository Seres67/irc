// mautrix-irc - A Matrix-IRC puppeting bridge.
// Copyright (C) 2025 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package ircfmt

import (
	"context"
	"fmt"
	"strings"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

var htmlParser = format.HTMLParser{
	TabsToSpaces:           4,
	Newline:                "\n",
	HorizontalLine:         "---",
	BoldConverter:          formattingAdder(bold),
	ItalicConverter:        formattingAdder(italic),
	StrikethroughConverter: formattingAdder(strikethrough),
	UnderlineConverter:     formattingAdder(underline),
	SpoilerConverter: func(text, reason string, ctx format.Context) string {
		return doAddFormatting(text, color+"01,01")
	},
	ColorConverter: func(text, fg, bg string, ctx format.Context) string {
		ircFG := reverseColors[strings.ToLower(fg)]
		ircBG := reverseColors[strings.ToLower(bg)]
		var resultFmt string
		if ircFG != "" {
			resultFmt = color + ircFG
			if ircBG != "" {
				resultFmt += "," + ircBG
			}
		}
		if len(fg) == 7 && fg[0] == '#' && isHex(fg[1:]) {
			resultFmt = hexColor + strings.ToUpper(fg[1:])
		}
		if resultFmt == "" {
			return text
		}
		return doAddFormatting(text, resultFmt)
	},
	MonospaceBlockConverter: func(code, language string, ctx format.Context) string {
		return doAddFormatting(code, monospace)
	},
	MonospaceConverter: formattingAdder(monospace),
	TextConverter: func(s string, context format.Context) string {
		return StripASCII(s)
	},
	PillConverter: func(displayname, mxid, eventID string, ctx format.Context) string {
		// const ContextKeyMentions = "_mentions"
		// switch {
		// //NOTE: im guessing that len(mxid) == 0 means that it comes from the non-matrix side
		// case len(mxid) == 0:
		// 	//TODO: could contain a user mention, convert to matrix?
		// 	//NOTE: how the fuck? search for colon? what
		// 	if false {
		// 		//TODO: change this condition to be when a mention has been found
		// 		existingMentions, _ := ctx.ReturnData[ContextKeyMentions].([]id.UserID)
		// 		ctx.ReturnData[ContextKeyMentions] = append(existingMentions, id.UserID(mxid))
		// 	}
		// 	return displayname
		// case len(eventID) > 0:
		// 	// Event ID link, always just show the link
		//
		// 	//NOTE: no clue what this is supposed to be
		// 	return fmt.Sprintf("https://matrix.to/#/%s/%s", mxid, eventID)
		// case mxid[0] == '#':
		// 	//NOTE: matrix -> irc
		// 	//TODO: what do you convert that to? can you even link to other channels in irc?
		// 	return mxid
		// case mxid[0] == '!':
		// 	//TODO: actual matrix room ID to convert to irc, send link?
		// 	return fmt.Sprintf("https://matrix.to/#/%s", mxid)
		// case mxid[0] == '@':
		// 	//TODO: user mentioned from matrix, convert to irc
		// 	return displayname
		// default:
		// 	// Other link (e.g. room ID link with display text), show text and link
		// 	return fmt.Sprintf("%s (https://matrix.to/#/%s)", displayname, mxid)
		// }

		if mxid[0] == '@' {
			return displayname
		}
		return format.DefaultPillConverter(displayname, mxid, eventID, ctx)

	},
}

func doAddFormatting(s, fmt string) string {
	return fmt + strings.ReplaceAll(strings.TrimRight(s, reset), reset, reset+fmt) + reset
}

func formattingAdder(fmt string) func(s string, context format.Context) string {
	return func(s string, context format.Context) string {
		return doAddFormatting(s, fmt)
	}
}

func ContentToASCII(ctx context.Context, content *event.MessageEventContent) string {
	if content.MsgType.IsMedia() && content.Body == content.GetFileName() {
		return ""
	} else if content.Format != event.FormatHTML {
		return StripASCII(content.Body)
	}
	return ParseHTML(ctx, content.FormattedBody)
}

func ParseHTML(ctx context.Context, html string) string {
	return htmlParser.Parse(html, format.NewContext(ctx))
}
