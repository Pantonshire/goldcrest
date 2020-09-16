package twitter1

import (
  "fmt"
  "github.com/golang/protobuf/ptypes"
  "goldcrest/rpc"
  "goldcrest/twitter1/model"
  "time"
)

func encodeAuthPair(secret, auth Auth) *rpc.Authentication {
  return &rpc.Authentication{
    ConsumerKey: auth.Key,
    AccessToken: auth.Token,
    SecretKey:   secret.Key,
    SecretToken: secret.Token,
  }
}

func decodeAuthPair(authMessage *rpc.Authentication) (secret Auth, auth Auth) {
  if authMessage == nil {
    return Auth{}, Auth{}
  }
  secret = Auth{Key: authMessage.SecretKey, Token: authMessage.SecretToken}
  auth = Auth{Key: authMessage.ConsumerKey, Token: authMessage.AccessToken}
  return secret, auth
}

func encodeTweetOptions(params TweetParams) *rpc.TweetOptions {
  return &rpc.TweetOptions{
    TrimUser:          params.TrimUser,
    IncludeMyRetweet:  params.IncludeMyRetweet,
    IncludeEntities:   params.IncludeEntities,
    IncludeExtAltText: params.IncludeExtAltText,
    IncludeCardUri:    params.TrimUser,
    Mode:              encodeTweetMode(params.Mode),
  }
}

func decodeTweetOptions(optsMessage *rpc.TweetOptions) TweetParams {
  if optsMessage == nil {
    return TweetParams{}
  }
  return TweetParams{
    TrimUser:          optsMessage.TrimUser,
    IncludeMyRetweet:  optsMessage.IncludeMyRetweet,
    IncludeEntities:   optsMessage.IncludeEntities,
    IncludeExtAltText: optsMessage.IncludeExtAltText,
    IncludeCardURI:    optsMessage.TrimUser,
    Mode:              decodeTweetMode(optsMessage.Mode),
  }
}

func encodeTweetMode(mode TweetMode) rpc.TweetOptions_TweetMode {
  if mode == ExtendedMode {
    return rpc.TweetOptions_EXTENDED
  }
  return rpc.TweetOptions_COMPAT
}

func decodeTweetMode(mode rpc.TweetOptions_TweetMode) TweetMode {
  if mode == rpc.TweetOptions_EXTENDED {
    return ExtendedMode
  }
  return CompatibilityMode
}

func tweetModelToMessage(mod model.Tweet) (*rpc.Tweet, error) {
  var err error
  var msg rpc.Tweet

  msg.Id = mod.ID

  if msg.CreatedAt, err = ptypes.TimestampProto(time.Time(mod.CreatedAt)); err != nil {
    return nil, err
  }

  if mod.ExtendedTweet != nil {
    msg.Text = mod.ExtendedTweet.FullText
  } else if mod.FullText != "" {
    msg.Text = mod.FullText
  } else {
    msg.Text = mod.Text
  }

  var displayTextRange []uint
  if mod.ExtendedTweet != nil {
    displayTextRange = mod.ExtendedTweet.DisplayTextRange
  } else {
    displayTextRange = mod.DisplayTextRange
  }
  if msg.TextDisplayRange, err = newIndicesMessage(displayTextRange); err != nil {
    return nil, err
  }

  msg.Truncated = mod.Truncated

  msg.Source = mod.Source

  if msg.User, err = userModelToMessage(mod.User); err != nil {
    return nil, err
  }

  if mod.ReplyStatusID != nil && mod.ReplyUserID != nil && mod.ReplyUserScreenName != nil {
    msg.Reply = &rpc.Tweet_RepliedTweet{RepliedTweet: &rpc.Tweet_Reply{
      ReplyToTweetId:    *mod.ReplyStatusID,
      ReplyToUserId:     *mod.ReplyUserID,
      ReplyToUserHandle: *mod.ReplyUserScreenName,
    }}
  } else {
    msg.Reply = &rpc.Tweet_NoReply{NoReply: true}
  }

  if mod.QuotedStatus != nil {
    quote, err := tweetModelToMessage(*mod.QuotedStatus)
    if err != nil {
      return nil, err
    }
    msg.Quote = &rpc.Tweet_QuotedTweet{QuotedTweet: quote}
  } else {
    msg.Quote = &rpc.Tweet_NoQuote{NoQuote: true}
  }

  if mod.RetweetedStatus != nil {
    retweet, err := tweetModelToMessage(*mod.RetweetedStatus)
    if err != nil {
      return nil, err
    }
    msg.Retweet = &rpc.Tweet_RetweetedTweet{RetweetedTweet: retweet}
  } else {
    msg.Retweet = &rpc.Tweet_NoRetweet{NoRetweet: true}
  }

  if mod.QuoteCount != nil {
    msg.QuoteCount = uint32(*mod.QuoteCount)
  }

  msg.ReplyCount = uint32(mod.ReplyCount)

  msg.RetweetCount = uint32(mod.RetweetCount)

  if mod.FavoriteCount != nil {
    msg.FavoriteCount = uint32(*mod.FavoriteCount)
  }

  if mod.CurrentUserRetweet != nil {
    msg.CurrentUserRetweetId = mod.CurrentUserRetweet.ID
  }

  var entities model.TweetEntities
  var extendedEntities model.TweetExtendedEntities

  if mod.ExtendedTweet != nil {
    entities, extendedEntities = mod.ExtendedTweet.Entities, mod.ExtendedTweet.ExtendedEntities
  } else {
    entities, extendedEntities = mod.Entities, mod.ExtendedEntities
  }

  if msg.Hashtags, err = symbolModelsToMessages(entities.Hashtags); err != nil {
    return nil, err
  }

  if msg.Urls, err = urlModelsToMessages(entities.URLs); err != nil {
    return nil, err
  }

  if msg.Mentions, err = mentionModelsToMessages(entities.Mentions); err != nil {
    return nil, err
  }

  if msg.Symbols, err = symbolModelsToMessages(entities.Symbols); err != nil {
    return nil, err
  }

  if msg.Polls, err = pollModelsToMessages(entities.Polls); err != nil {
    return nil, err
  }

  var media []model.Media
  mediaIDs := make(map[uint64]bool)
  for _, mm := range extendedEntities.Media {
    if !mediaIDs[mm.ID] {
      media = append(media, mm)
      mediaIDs[mm.ID] = true
    }
  }
  for _, mm := range entities.Media {
    if !mediaIDs[mm.ID] {
      media = append(media, mm)
      mediaIDs[mm.ID] = true
    }
  }
  if msg.Media, err = mediaModelsToMessages(media); err != nil {
    return nil, err
  }

  return &msg, nil
}

func decodeTweet(msg *rpc.Tweet) Tweet {
  if msg == nil {
    return Tweet{}
  }
  tweet := Tweet{
    ID:                   msg.Id,
    CreatedAt:            msg.CreatedAt.AsTime(),
    Text:                 msg.Text,
    TextDisplayRange:     decodeIndices(msg.TextDisplayRange),
    Truncated:            msg.Truncated,
    Source:               msg.Source,
    User:                 decodeUser(msg.User),
    Quotes:               uint(msg.QuoteCount),
    Replies:              uint(msg.ReplyCount),
    Retweets:             uint(msg.RetweetCount),
    Likes:                uint(msg.FavoriteCount),
    CurrentUserLiked:     msg.Favorited,
    CurrentUserRetweeted: msg.Retweeted,
    Hashtags:             decodeSymbols(msg.Hashtags),
    URLs:                 decodeURLs(msg.Urls),
    Mentions:             decodeMentions(msg.Mentions),
    Symbols:              decodeSymbols(msg.Symbols),
    Media:                decodeMedia(msg.Media),
    Polls:                decodePolls(msg.Polls),
    PossiblySensitive:    msg.PossiblySensitive,
    FilterLevel:          msg.FilterLevel,
    Lang:                 msg.Lang,
    WithheldCopyright:    msg.WithheldCopyright,
    WithheldCounties:     msg.WithheldCountries,
    WithheldScope:        msg.WithheldScope,
  }
  if msg.Reply != nil {
    if reply, ok := msg.Reply.(*rpc.Tweet_RepliedTweet); ok {
      if reply.RepliedTweet != nil {
        tweet.RepliedTo = &ReplyData{
          TweetID:    reply.RepliedTweet.ReplyToTweetId,
          UserID:     reply.RepliedTweet.ReplyToUserId,
          UserHandle: reply.RepliedTweet.ReplyToUserHandle,
        }
      }
    }
  }
  if msg.Quote != nil {
    if quote, ok := msg.Quote.(*rpc.Tweet_QuotedTweet); ok {
      decodedQuote := decodeTweet(quote.QuotedTweet)
      tweet.Quoted = &decodedQuote
    }
  }
  if msg.Retweet != nil {
    if retweet, ok := msg.Retweet.(*rpc.Tweet_RetweetedTweet); ok {
      decodedRetweet := decodeTweet(retweet.RetweetedTweet)
      tweet.Retweeted = &decodedRetweet
    }
  }
  if msg.CurrentUserRetweetId != 0 {
    retweetID := msg.CurrentUserRetweetId
    tweet.CurrentUserRetweetID = &retweetID
  }
  return tweet
}

func userModelToMessage(mod model.User) (*rpc.User, error) {
  var err error
  var msg rpc.User

  msg.Id = mod.ID

  msg.Handle = mod.ScreenName

  msg.DisplayName = mod.Name

  if msg.CreatedAt, err = ptypes.TimestampProto(time.Time(mod.CreatedAt)); err != nil {
    return nil, err
  }

  if mod.Description != nil {
    msg.Bio = *mod.Description
  }

  if mod.URL != nil {
    msg.Url = *mod.URL
  }

  if mod.Location != nil {
    msg.Location = *mod.Location
  }

  msg.Protected = mod.Protected

  msg.Verified = mod.Verified

  msg.FollowerCount = uint32(mod.FollowersCount)

  msg.FollowingCount = uint32(mod.FriendsCount)

  msg.ListedCount = uint32(mod.ListedCount)

  msg.FavoritesCount = uint32(mod.FavoritesCount)

  msg.StatusesCount = uint32(mod.StatusesCount)

  msg.ProfileBanner = mod.ProfileBanner

  msg.ProfileImage = mod.ProfileImage

  msg.DefaultProfile = mod.DefaultProfile

  msg.DefaultProfileImage = mod.DefaultProfileImage

  msg.WithheldCountries = mod.WithheldCountries

  if mod.WithheldScope != nil {
    msg.WithheldScope = *mod.WithheldScope
  }

  if msg.UrlUrls, err = urlModelsToMessages(mod.Entities.URL.URLs); err != nil {
    return nil, err
  }

  if msg.BioUrls, err = urlModelsToMessages(mod.Entities.Description.URLs); err != nil {
    return nil, err
  }

  return &msg, nil
}

func decodeUser(msg *rpc.User) User {
  if msg == nil {
    return User{}
  }
  return User{
    ID:                  msg.Id,
    Handle:              msg.Handle,
    DisplayName:         msg.DisplayName,
    CreatedAt:           msg.CreatedAt.AsTime(),
    Bio:                 msg.Bio,
    URL:                 msg.Url,
    Location:            msg.Location,
    Protected:           msg.Protected,
    Verified:            msg.Verified,
    FollowerCount:       uint(msg.FollowerCount),
    FollowingCount:      uint(msg.FollowingCount),
    ListedCount:         uint(msg.ListedCount),
    FavouritesCount:     uint(msg.FavoritesCount),
    StatusesCount:       uint(msg.StatusesCount),
    ProfileBanner:       msg.ProfileBanner,
    ProfileImage:        msg.ProfileImage,
    DefaultProfile:      msg.DefaultProfile,
    DefaultProfileImage: msg.DefaultProfileImage,
    WithheldCountries:   msg.WithheldCountries,
    WithheldScope:       msg.WithheldScope,
    URLs:                decodeURLs(msg.UrlUrls),
    BioURLs:             decodeURLs(msg.BioUrls),
  }
}

func newIndicesMessage(indices []uint) (*rpc.Indices, error) {
  if len(indices) != 2 {
    return nil, fmt.Errorf("expected [start,end] index values pair, got %v", indices)
  }
  return &rpc.Indices{Start: uint32(indices[0]), End: uint32(indices[1])}, nil
}

func decodeIndices(msg *rpc.Indices) Indices {
  if msg == nil {
    return Indices{}
  }
  return Indices{Start: uint(msg.Start), End: uint(msg.End)}
}

func urlModelToMessage(mod model.URL) (*rpc.URL, error) {
  var err error
  msg := rpc.URL{
    TwitterUrl:  mod.URL,
    DisplayUrl:  mod.DisplayURL,
    ExpandedUrl: mod.ExpandedURL,
  }
  if msg.Indices, err = newIndicesMessage(mod.Indices); err != nil {
    return nil, err
  }
  return &msg, nil
}

func urlModelsToMessages(mods []model.URL) ([]*rpc.URL, error) {
  var err error
  msgs := make([]*rpc.URL, len(mods))
  for i, mod := range mods {
    if msgs[i], err = urlModelToMessage(mod); err != nil {
      return nil, err
    }
  }
  return msgs, nil
}

func decodeURL(msg *rpc.URL) URL {
  if msg == nil {
    return URL{}
  }
  return URL{
    Indices:     decodeIndices(msg.Indices),
    TwitterURL:  msg.TwitterUrl,
    DisplayURL:  msg.DisplayUrl,
    ExpandedURL: msg.ExpandedUrl,
  }
}

func decodeURLs(msgs []*rpc.URL) []URL {
  urls := make([]URL, len(msgs))
  for i, msg := range msgs {
    urls[i] = decodeURL(msg)
  }
  return urls
}

func symbolModelsToMessages(mods []model.Symbol) ([]*rpc.Symbol, error) {
  var err error
  msgs := make([]*rpc.Symbol, len(mods))
  for i, mod := range mods {
    msg := rpc.Symbol{Text: mod.Text}
    if msg.Indices, err = newIndicesMessage(mod.Indices); err != nil {
      return nil, err
    }
    msgs[i] = &msg
  }
  return msgs, nil
}

func decodeSymbols(msgs []*rpc.Symbol) []Symbol {
  symbols := make([]Symbol, len(msgs))
  for i, msg := range msgs {
    if msg != nil {
      symbols[i] = Symbol{
        Indices: decodeIndices(msg.Indices),
        Text:    msg.Text,
      }
    }
  }
  return symbols
}

func mentionModelsToMessages(mods []model.Mention) ([]*rpc.Mention, error) {
  var err error
  msgs := make([]*rpc.Mention, len(mods))
  for i, mod := range mods {
    msg := rpc.Mention{
      UserId:      mod.ID,
      Handle:      mod.ScreenName,
      DisplayName: mod.Name,
    }
    if msg.Indices, err = newIndicesMessage(mod.Indices); err != nil {
      return nil, err
    }
    msgs[i] = &msg
  }
  return msgs, nil
}

func decodeMentions(msgs []*rpc.Mention) []Mention {
  mentions := make([]Mention, len(msgs))
  for i, msg := range msgs {
    if msg != nil {
      mentions[i] = Mention{
        Indices:         decodeIndices(msg.Indices),
        UserID:          msg.UserId,
        UserHandle:      msg.Handle,
        UserDisplayName: msg.DisplayName,
      }
    }
  }
  return mentions
}

func mediaModelsToMessages(mods []model.Media) ([]*rpc.Media, error) {
  var err error
  msgs := make([]*rpc.Media, len(mods))
  for i, mod := range mods {
    msg := rpc.Media{
      Id:   mod.ID,
      Type: mod.Type,
      Alt:  mod.AltText,
      Thumb: &rpc.Media_MediaSize{
        Width:  uint32(mod.Sizes.Thumb.W),
        Height: uint32(mod.Sizes.Thumb.H),
        Resize: mod.Sizes.Thumb.Resize,
      },
      Small: &rpc.Media_MediaSize{
        Width:  uint32(mod.Sizes.Small.W),
        Height: uint32(mod.Sizes.Small.H),
        Resize: mod.Sizes.Small.Resize,
      },
      Medium: &rpc.Media_MediaSize{
        Width:  uint32(mod.Sizes.Medium.W),
        Height: uint32(mod.Sizes.Medium.H),
        Resize: mod.Sizes.Medium.Resize,
      },
      Large: &rpc.Media_MediaSize{
        Width:  uint32(mod.Sizes.Large.W),
        Height: uint32(mod.Sizes.Large.H),
        Resize: mod.Sizes.Large.Resize,
      },
    }
    if msg.Url, err = urlModelToMessage(mod.URL); err != nil {
      return nil, err
    }
    if mod.MediaURLHttps != "" {
      msg.MediaUrl = mod.MediaURLHttps
    } else {
      msg.MediaUrl = mod.MediaURL
    }
    if mod.SourceStatusID != nil {
      msg.Source = &rpc.Media_SourceTweetId{SourceTweetId: *mod.SourceStatusID}
    } else {
      msg.Source = &rpc.Media_NoSource{NoSource: true}
    }
    msgs[i] = &msg
  }
  return msgs, nil
}

func decodeMedia(msgs []*rpc.Media) []Media {
  media := make([]Media, len(msgs))
  for i, msg := range msgs {
    if msg != nil {
      media[i] = Media{
        URL:           decodeURL(msg.Url),
        ID:            msg.Id,
        Type:          msg.Type,
        MediaURL:      msg.MediaUrl,
        Alt:           msg.Alt,
        SourceTweetID: decodeMediaSource(msg),
        Thumb:         decodeMediaSize(msg.Thumb),
        Small:         decodeMediaSize(msg.Small),
        Medium:        decodeMediaSize(msg.Medium),
        Large:         decodeMediaSize(msg.Large),
      }
    }
  }
  return media
}

func decodeMediaSource(msg *rpc.Media) *uint64 {
  if msg.Source == nil {
    return nil
  }
  if source, ok := msg.Source.(*rpc.Media_SourceTweetId); ok {
    sourceID := source.SourceTweetId
    return &sourceID
  }
  return nil
}

func decodeMediaSize(msg *rpc.Media_MediaSize) MediaSize {
  if msg == nil {
    return MediaSize{}
  }
  return MediaSize{Width: uint(msg.Width), Height: uint(msg.Height), Resize: msg.Resize}
}

func pollModelsToMessages(mods []model.Poll) ([]*rpc.Poll, error) {
  var err error
  msgs := make([]*rpc.Poll, len(mods))
  for i, mod := range mods {
    msg := rpc.Poll{DurationMinutes: uint32(mod.DurationMinutes)}
    if msg.EndTime, err = ptypes.TimestampProto(time.Time(mod.EndTime)); err != nil {
      return nil, err
    }
    msg.Options = make([]*rpc.Poll_PollOption, len(mod.Options))
    for j, optionMod := range mod.Options {
      msg.Options[j] = &rpc.Poll_PollOption{
        Position: uint32(optionMod.Position),
        Text:     optionMod.Text,
      }
    }
    msgs[i] = &msg
  }
  return msgs, nil
}

func decodePolls(msgs []*rpc.Poll) []Poll {
  polls := make([]Poll, len(msgs))
  for i, msg := range msgs {
    if msg != nil {
      polls[i] = Poll{
        EndTime:  msg.EndTime.AsTime(),
        Duration: time.Minute * time.Duration(msg.DurationMinutes),
        Options:  make([]PollOption, len(msg.Options)),
      }
      for j, optMsg := range msg.Options {
        if optMsg != nil {
          polls[i].Options[j] = PollOption{
            Position: uint(optMsg.Position),
            Text:     optMsg.Text,
          }
        }
      }
    }
  }
  return polls
}