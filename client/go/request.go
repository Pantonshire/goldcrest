package goldcrest

import (
  pb "github.com/pantonshire/goldcrest/protocol"
  "time"
)

type authentication struct {
  consumerKey, accessToken, secretKey, secretToken string
}

func (auth authentication) ser() *pb.Authentication {
  return &pb.Authentication{
    ConsumerKey: auth.consumerKey,
    AccessToken: auth.accessToken,
    SecretKey:   auth.secretKey,
    SecretToken: auth.secretToken,
  }
}

type TweetMode uint8

const (
  CompatibilityMode TweetMode = iota
  ExtendedMode
)

func (m TweetMode) ser() pb.TweetOptions_Mode {
  switch m {
  case ExtendedMode:
    return pb.TweetOptions_EXTENDED
  default:
    return pb.TweetOptions_COMPAT
  }
}

type TweetOptions struct {
  trimUser          bool
  includeMyRetweet  bool
  includeEntities   bool
  includeExtAltText bool
  includeCardURI    bool
  mode              TweetMode
}

func NewTweetOptions() TweetOptions {
  return TweetOptions{
    trimUser:          false,
    includeMyRetweet:  true,
    includeEntities:   true,
    includeExtAltText: true,
    includeCardURI:    true,
    mode:              ExtendedMode,
  }
}

func (opts TweetOptions) WithTrimUser(b bool) TweetOptions {
  opts.trimUser = b
  return opts
}

func (opts TweetOptions) WithMyRetweet(b bool) TweetOptions {
  opts.includeMyRetweet = b
  return opts
}

func (opts TweetOptions) WithEntities(b bool) TweetOptions {
  opts.includeEntities = b
  return opts
}

func (opts TweetOptions) WithAltText(b bool) TweetOptions {
  opts.includeExtAltText = b
  return opts
}

func (opts TweetOptions) WithCardURI(b bool) TweetOptions {
  opts.includeCardURI = b
  return opts
}

func (opts TweetOptions) WithMode(m TweetMode) TweetOptions {
  opts.mode = m
  return opts
}

func (opts TweetOptions) ser() *pb.TweetOptions {
  return &pb.TweetOptions{
    TrimUser:          opts.trimUser,
    IncludeMyRetweet:  opts.includeMyRetweet,
    IncludeEntities:   opts.includeEntities,
    IncludeExtAltText: opts.includeExtAltText,
    IncludeCardUri:    opts.includeCardURI,
    Mode:              opts.mode.ser(),
  }
}

type TimelineOptions struct {
  count    uint
  min, max *uint64
}

func NewTimelineOptions(count uint) TimelineOptions {
  return TimelineOptions{count: count}
}

func (tlopts TimelineOptions) WithMin(min uint64) TimelineOptions {
  tlopts.min = new(uint64)
  *tlopts.min = min
  return tlopts
}

func (tlopts TimelineOptions) WithMax(max uint64) TimelineOptions {
  tlopts.max = new(uint64)
  *tlopts.max = max
  return tlopts
}

func (tlopts TimelineOptions) ser(twopts TweetOptions) *pb.TimelineOptions {
  msg := pb.TimelineOptions{
    Count:  uint32(tlopts.count),
    Twopts: twopts.ser(),
  }
  if tlopts.min != nil {
    msg.MinId = &pb.OptFixed64{Val: *tlopts.min}
  }
  if tlopts.max != nil {
    msg.MaxId = &pb.OptFixed64{Val: *tlopts.max}
  }
  return &msg
}

type SearchResultType uint8

const (
  SearchMixed SearchResultType = iota
  SearchRecent
  SearchPopular
)

func (t SearchResultType) ser() pb.SearchRequest_ResultType {
  switch t {
  case SearchRecent:
    return pb.SearchRequest_RECENT
  case SearchPopular:
    return pb.SearchRequest_POPULAR
  default:
    return pb.SearchRequest_MIXED
  }
}

type SearchOptions struct {
  query             string
  geo, lang, locale *string
  resType           SearchResultType
  until             *time.Time
}

func NewSearchOptions(query string) SearchOptions {
  return SearchOptions{
    query:   query,
    resType: SearchMixed,
  }
}

func (opts SearchOptions) WithGeocode(geocode string) SearchOptions {
  opts.geo = new(string)
  *opts.geo = geocode
  return opts
}

func (opts SearchOptions) WithLang(lang string) SearchOptions {
  opts.lang = new(string)
  *opts.lang = lang
  return opts
}

func (opts SearchOptions) WithLocale(locale string) SearchOptions {
  opts.locale = new(string)
  *opts.locale = locale
  return opts
}

func (opts SearchOptions) WithResultType(resType SearchResultType) SearchOptions {
  opts.resType = resType
  return opts
}

func (opts SearchOptions) WithUntilTime(until time.Time) SearchOptions {
  opts.until = new(time.Time)
  *opts.until = until
  return opts
}

func serSearchRequest(auth authentication, searchOpts SearchOptions, twOpts TweetOptions, tlOpts TimelineOptions) *pb.SearchRequest {
  req := pb.SearchRequest{
    Auth:            auth.ser(),
    Query:           searchOpts.query,
    ResultType:      searchOpts.resType.ser(),
    TimelineOptions: tlOpts.ser(twOpts),
  }
  if searchOpts.geo != nil {
    req.Geocode = &pb.OptString{Val: *searchOpts.geo}
  }
  if searchOpts.lang != nil {
    req.Lang = &pb.OptString{Val: *searchOpts.lang}
  }
  if searchOpts.locale != nil {
    req.Locale = &pb.OptString{Val: *searchOpts.locale}
  }
  if searchOpts.until != nil {
    req.UntilTimestamp = &pb.OptInt64{Val: searchOpts.until.Unix()}
  }
  return &req
}

type UserIdentifier interface {
  serIntoUserTimelineRequest(req *pb.UserTimelineRequest)
}

type userIdentifierID uint64

func UserID(id uint64) UserIdentifier {
  return userIdentifierID(id)
}

func (uid userIdentifierID) serIntoUserTimelineRequest(req *pb.UserTimelineRequest) {
  req.User = &pb.UserTimelineRequest_UserId{UserId: uint64(uid)}
}

type userIdentifierHandle string

func UserHandle(handle string) UserIdentifier {
  return userIdentifierHandle(handle)
}

func (uid userIdentifierHandle) serIntoUserTimelineRequest(req *pb.UserTimelineRequest) {
  req.User = &pb.UserTimelineRequest_UserHandle{UserHandle: string(uid)}
}

type TweetComposer struct {
  text              string
  replyID           *uint64
  excludeUserIDs    []uint64
  attachmentURL     *string
  mediaIDs          []uint64
  possiblySensitive bool
  enableDMCommands  bool
  failDMCommands    bool
}

func NewTweetComposer(text string) TweetComposer {
  return TweetComposer{
    text: text,
  }
}

func (com TweetComposer) ReplyTo(tweetID uint64, excludeUserIDs ...uint64) TweetComposer {
  com.replyID = new(uint64)
  *com.replyID = tweetID
  com.excludeUserIDs = make([]uint64, len(excludeUserIDs))
  copy(com.excludeUserIDs, excludeUserIDs)
  return com
}

func (com TweetComposer) WithAttachment(url string) TweetComposer {
  com.attachmentURL = new(string)
  *com.attachmentURL = url
  return com
}

func (com TweetComposer) WithMedia(ids ...uint64) TweetComposer {
  com.mediaIDs = make([]uint64, len(ids))
  copy(com.mediaIDs, ids)
  return com
}

func (com TweetComposer) WithSensitive(sensitive bool) TweetComposer {
  com.possiblySensitive = sensitive
  return com
}

func (com TweetComposer) WithEnableDMCommands(enabled bool) TweetComposer {
  com.enableDMCommands = enabled
  return com
}

func (com TweetComposer) WithFailDMCommands(fail bool) TweetComposer {
  com.failDMCommands = fail
  return com
}

func (com TweetComposer) ser(auth authentication, twopts TweetOptions) *pb.PublishTweetRequest {
  req := pb.PublishTweetRequest{
    Auth:              auth.ser(),
    Text:              com.text,
    MediaIds:          com.mediaIDs,
    PossiblySensitive: com.possiblySensitive,
    EnableDmCommands:  com.enableDMCommands,
    FailDmCommands:    com.failDMCommands,
    Twopts:            twopts.ser(),
  }
  if com.replyID != nil {
    req.ReplyId = &pb.OptFixed64{Val: *com.replyID}
    req.AutoPopulateReplyMetadata = true
    if com.excludeUserIDs != nil {
      req.ExcludeReplyUserIds = com.excludeUserIDs
    }
  }
  if com.attachmentURL != nil {
    req.AttachmentUrl = &pb.OptString{Val: *com.attachmentURL}
  }
  return &req
}

type ProfileUpdater struct {
  name             *string
  url              *string
  location         *string
  bio              *string
  profileLinkColor *string
}

func NewProfileUpdater() ProfileUpdater {
  return ProfileUpdater{}
}

func (pu ProfileUpdater) WithName(name string) ProfileUpdater {
  pu.name = new(string)
  *pu.name = name
  return pu
}

func (pu ProfileUpdater) WithURL(url string) ProfileUpdater {
  pu.url = new(string)
  *pu.url = url
  return pu
}

func (pu ProfileUpdater) WithLocation(location string) ProfileUpdater {
  pu.location = new(string)
  *pu.location = location
  return pu
}

func (pu ProfileUpdater) WithBio(bio string) ProfileUpdater {
  pu.bio = new(string)
  *pu.bio = bio
  return pu
}

func (pu ProfileUpdater) WithProfileLinkColor(color string) ProfileUpdater {
  pu.profileLinkColor = new(string)
  *pu.profileLinkColor = color
  return pu
}

func (pu ProfileUpdater) ser(auth authentication, includeEntities, includeStatuses bool) *pb.UpdateProfileRequest {
  req := pb.UpdateProfileRequest{
    Auth:            auth.ser(),
    IncludeEntities: includeEntities,
    IncludeStatuses: includeStatuses,
  }
  if pu.name != nil {
    req.Name = &pb.OptString{Val: *pu.name}
  }
  if pu.url != nil {
    req.Url = &pb.OptString{Val: *pu.url}
  }
  if pu.location != nil {
    req.Location = &pb.OptString{Val: *pu.location}
  }
  if pu.bio != nil {
    req.Bio = &pb.OptString{Val: *pu.bio}
  }
  if pu.profileLinkColor != nil {
    req.LinkColor = &pb.OptString{Val: *pu.profileLinkColor}
  }
  return &req
}
