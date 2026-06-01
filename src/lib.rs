//! Async YouTube video URL extraction and downloading.
//!
//! This crate intentionally focuses on the small feature set requested here:
//! resolving direct media URLs for a public YouTube video id and downloading a
//! playable muxed format with Tokio + Reqwest. YouTube changes its private
//! Innertube API regularly, so the implementation keeps the protocol boundary
//! isolated in [`YoutubeClient::player_response`].

use std::{path::Path, sync::Arc};

use futures_util::StreamExt;
use regex::Regex;
use reqwest::{header, Url};
use serde::{Deserialize, Serialize};
use tokio::{fs::File, io::AsyncWriteExt};

const DEFAULT_INNERTUBE_BASE_URL: &str = "https://www.youtube.com/youtubei/v1";
const DEFAULT_INNERTUBE_API_KEY: &str = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8";
const DEFAULT_ANDROID_CLIENT_VERSION: &str = "19.09.37";

/// Result alias used by this crate.
pub type Result<T> = std::result::Result<T, YoutubeError>;

/// Errors returned by URL extraction and downloading operations.
#[derive(Debug, thiserror::Error)]
pub enum YoutubeError {
    #[error("invalid YouTube video id or URL: {0}")]
    InvalidVideoId(String),

    #[error("the video is unavailable: {0}")]
    VideoUnavailable(String),

    #[error("no downloadable muxed video format was returned")]
    NoDownloadableFormat,

    #[error("format {itag} requires signature deciphering, which is not implemented")]
    SignatureCipherRequired { itag: u32 },

    #[error("invalid URL: {0}")]
    InvalidUrl(#[from] url::ParseError),

    #[error("http request failed: {0}")]
    Request(#[from] reqwest::Error),

    #[error("file I/O failed: {0}")]
    Io(#[from] std::io::Error),

    #[error("failed to parse YouTube response: {0}")]
    Json(#[from] serde_json::Error),
}

/// Async client for extracting and downloading YouTube streams.
#[derive(Clone, Debug)]
pub struct YoutubeClient {
    http: reqwest::Client,
    innertube_base_url: Url,
    innertube_api_key: Arc<str>,
    client_version: Arc<str>,
}

impl YoutubeClient {
    /// Create a client configured for YouTube's Innertube player endpoint.
    pub fn new() -> Result<Self> {
        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::USER_AGENT,
            header::HeaderValue::from_static(
                "com.google.android.youtube/19.09.37 (Linux; U; Android 11)",
            ),
        );

        let http = reqwest::Client::builder()
            .default_headers(headers)
            .redirect(reqwest::redirect::Policy::limited(10))
            .build()?;

        Self::with_http_client(http)
    }

    /// Create a client with a caller-provided [`reqwest::Client`].
    pub fn with_http_client(http: reqwest::Client) -> Result<Self> {
        Ok(Self {
            http,
            innertube_base_url: Url::parse(DEFAULT_INNERTUBE_BASE_URL)?,
            innertube_api_key: DEFAULT_INNERTUBE_API_KEY.into(),
            client_version: DEFAULT_ANDROID_CLIENT_VERSION.into(),
        })
    }

    /// Override the Innertube base URL.
    ///
    /// This is mainly useful for tests that run against a local mock server.
    pub fn with_innertube_base_url(mut self, base_url: impl AsRef<str>) -> Result<Self> {
        self.innertube_base_url = Url::parse(base_url.as_ref())?;
        Ok(self)
    }

    /// Override the Innertube API key.
    pub fn with_api_key(mut self, api_key: impl Into<Arc<str>>) -> Self {
        self.innertube_api_key = api_key.into();
        self
    }

    /// Extract direct stream URLs for a YouTube video id or watch URL.
    pub async fn extract_urls(&self, video_id_or_url: &str) -> Result<Vec<VideoFormat>> {
        let video_id = parse_video_id(video_id_or_url)?;
        let response = self.player_response(&video_id).await?;

        let playability = response.playability_status;
        if !playability.is_ok() {
            return Err(YoutubeError::VideoUnavailable(
                playability.reason.unwrap_or(playability.status),
            ));
        }

        let Some(streaming_data) = response.streaming_data else {
            return Err(YoutubeError::NoDownloadableFormat);
        };

        let mut formats = streaming_data.formats;
        formats.extend(streaming_data.adaptive_formats);

        formats
            .into_iter()
            .map(VideoFormat::try_from)
            .collect::<Result<Vec<_>>>()
    }

    /// Download the best muxed format for a YouTube video id or watch URL.
    ///
    /// "Muxed" means the selected stream contains both audio and video. Adaptive
    /// streams are exposed by [`extract_urls`](Self::extract_urls), but are not
    /// selected here because they require a separate muxing step after download.
    pub async fn download_video(
        &self,
        video_id_or_url: &str,
        output_path: impl AsRef<Path>,
    ) -> Result<DownloadedVideo> {
        let formats = self.extract_urls(video_id_or_url).await?;
        let format = formats
            .into_iter()
            .filter(VideoFormat::is_muxed)
            .max_by_key(VideoFormat::sort_key)
            .ok_or(YoutubeError::NoDownloadableFormat)?;

        let response = self
            .http
            .get(format.url.clone())
            .send()
            .await?
            .error_for_status()?;
        let mut stream = response.bytes_stream();
        let mut file = File::create(output_path.as_ref()).await?;
        let mut bytes_written = 0_u64;

        while let Some(chunk) = stream.next().await {
            let chunk = chunk?;
            bytes_written += chunk.len() as u64;
            file.write_all(&chunk).await?;
        }
        file.flush().await?;

        Ok(DownloadedVideo {
            format,
            bytes_written,
        })
    }

    async fn player_response(&self, video_id: &str) -> Result<PlayerResponse> {
        let mut url = self.innertube_base_url.join("./player")?;
        url.query_pairs_mut()
            .append_pair("key", &self.innertube_api_key);

        let request = PlayerRequest::android(video_id, &self.client_version);
        let response = self
            .http
            .post(url)
            .json(&request)
            .send()
            .await?
            .error_for_status()?;

        Ok(response.json().await?)
    }
}

impl Default for YoutubeClient {
    fn default() -> Self {
        Self::new().expect("default YouTube client configuration must be valid")
    }
}

/// A direct media stream returned by YouTube.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct VideoFormat {
    pub itag: u32,
    pub mime_type: String,
    pub quality_label: Option<String>,
    pub bitrate: Option<u64>,
    pub width: Option<u32>,
    pub height: Option<u32>,
    pub content_length: Option<u64>,
    pub url: Url,
}

impl VideoFormat {
    /// Returns true when this format contains both audio and video.
    pub fn is_muxed(&self) -> bool {
        self.mime_type.starts_with("video/") && self.mime_type.contains("mp4a")
    }

    fn sort_key(&self) -> (u32, u64) {
        (
            self.height.unwrap_or_default(),
            self.bitrate.unwrap_or_default(),
        )
    }
}

/// Information about a completed download.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct DownloadedVideo {
    pub format: VideoFormat,
    pub bytes_written: u64,
}

impl TryFrom<RawFormat> for VideoFormat {
    type Error = YoutubeError;

    fn try_from(raw: RawFormat) -> Result<Self> {
        let url = match raw.url {
            Some(url) => Url::parse(&url)?,
            None if raw.signature_cipher.is_some() || raw.cipher.is_some() => {
                return Err(YoutubeError::SignatureCipherRequired { itag: raw.itag });
            }
            None => return Err(YoutubeError::NoDownloadableFormat),
        };

        Ok(Self {
            itag: raw.itag,
            mime_type: raw.mime_type,
            quality_label: raw.quality_label,
            bitrate: raw.bitrate,
            width: raw.width,
            height: raw.height,
            content_length: raw
                .content_length
                .as_deref()
                .and_then(|value| value.parse::<u64>().ok()),
            url,
        })
    }
}

/// Parse the canonical 11-character YouTube video id from either an id or URL.
pub fn parse_video_id(input: &str) -> Result<String> {
    let trimmed = input.trim();
    let id_pattern = Regex::new(r"^[A-Za-z0-9_-]{11}$").expect("video id regex is valid");
    if id_pattern.is_match(trimmed) {
        return Ok(trimmed.to_owned());
    }

    let url = Url::parse(trimmed).map_err(|_| YoutubeError::InvalidVideoId(trimmed.to_owned()))?;
    let host = url.host_str().unwrap_or_default();

    let candidate = if host.ends_with("youtube.com") || host.ends_with("youtube-nocookie.com") {
        if url.path() == "/watch" {
            url.query_pairs()
                .find_map(|(key, value)| (key == "v").then(|| value.into_owned()))
        } else if url.path().starts_with("/embed/") || url.path().starts_with("/shorts/") {
            url.path_segments()
                .and_then(|mut segments| segments.nth(1))
                .map(ToOwned::to_owned)
        } else {
            None
        }
    } else if host == "youtu.be" {
        url.path_segments()
            .and_then(|mut segments| segments.next())
            .map(ToOwned::to_owned)
    } else {
        None
    };

    candidate
        .filter(|value| id_pattern.is_match(value))
        .ok_or_else(|| YoutubeError::InvalidVideoId(trimmed.to_owned()))
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct PlayerRequest<'a> {
    context: RequestContext<'a>,
    video_id: &'a str,
    content_check_ok: bool,
    racy_check_ok: bool,
}

impl<'a> PlayerRequest<'a> {
    fn android(video_id: &'a str, client_version: &'a str) -> Self {
        Self {
            context: RequestContext {
                client: ClientContext {
                    client_name: "ANDROID",
                    client_version,
                    android_sdk_version: 30,
                    hl: "en",
                    gl: "US",
                },
            },
            video_id,
            content_check_ok: true,
            racy_check_ok: true,
        }
    }
}

#[derive(Debug, Serialize)]
struct RequestContext<'a> {
    client: ClientContext<'a>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct ClientContext<'a> {
    client_name: &'a str,
    client_version: &'a str,
    android_sdk_version: u32,
    hl: &'a str,
    gl: &'a str,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct PlayerResponse {
    playability_status: PlayabilityStatus,
    streaming_data: Option<StreamingData>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct PlayabilityStatus {
    status: String,
    reason: Option<String>,
}

impl PlayabilityStatus {
    fn is_ok(&self) -> bool {
        self.status.eq_ignore_ascii_case("OK")
    }
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct StreamingData {
    #[serde(default)]
    formats: Vec<RawFormat>,
    #[serde(default)]
    adaptive_formats: Vec<RawFormat>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct RawFormat {
    itag: u32,
    mime_type: String,
    quality_label: Option<String>,
    bitrate: Option<u64>,
    width: Option<u32>,
    height: Option<u32>,
    content_length: Option<String>,
    url: Option<String>,
    signature_cipher: Option<String>,
    cipher: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::SocketAddr;
    use tempfile::NamedTempFile;
    use tokio::{
        io::{AsyncReadExt, AsyncWriteExt},
        net::TcpListener,
    };

    #[test]
    fn parses_video_ids_from_supported_inputs() {
        let cases = [
            ("dQw4w9WgXcQ", "dQw4w9WgXcQ"),
            ("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"),
            ("https://youtu.be/dQw4w9WgXcQ?t=10", "dQw4w9WgXcQ"),
            ("https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
            ("https://www.youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ];

        for (input, expected) in cases {
            assert_eq!(parse_video_id(input).unwrap(), expected);
        }
    }

    #[test]
    fn rejects_invalid_video_ids() {
        assert!(matches!(
            parse_video_id("https://example.com/watch?v=dQw4w9WgXcQ"),
            Err(YoutubeError::InvalidVideoId(_))
        ));
        assert!(matches!(
            parse_video_id("too-short"),
            Err(YoutubeError::InvalidVideoId(_))
        ));
    }

    #[tokio::test]
    async fn extracts_urls_from_player_response() {
        let video_url = "https://cdn.example/video.mp4?ratebypass=yes";
        let server = TestServer::start(format!(
            r#"{{"playabilityStatus":{{"status":"OK"}},"streamingData":{{"formats":[{{"itag":18,"mimeType":"video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"","qualityLabel":"360p","bitrate":500000,"width":640,"height":360,"contentLength":"42","url":"{video_url}"}}],"adaptiveFormats":[]}}}}"#,
        ))
        .await;

        let client = test_client()
            .with_innertube_base_url(server.base_url())
            .unwrap()
            .with_api_key("test-key");

        let formats = client.extract_urls("dQw4w9WgXcQ").await.unwrap();

        assert_eq!(formats.len(), 1);
        assert_eq!(formats[0].itag, 18);
        assert_eq!(formats[0].height, Some(360));
        assert_eq!(formats[0].content_length, Some(42));
        assert_eq!(formats[0].url.as_str(), video_url);
    }

    #[tokio::test]
    async fn downloads_best_muxed_format() {
        let server = DownloadServer::start().await;
        let player_body = format!(
            r#"{{"playabilityStatus":{{"status":"OK"}},"streamingData":{{"formats":[{{"itag":18,"mimeType":"video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"","qualityLabel":"360p","bitrate":500000,"height":360,"url":"{}/video-low"}},{{"itag":22,"mimeType":"video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"","qualityLabel":"720p","bitrate":1500000,"height":720,"url":"{}/video-high"}}],"adaptiveFormats":[]}}}}"#,
            server.base_url().trim_end_matches('/'),
            server.base_url().trim_end_matches('/'),
        );
        server.set_player_body(player_body).await;

        let client = test_client()
            .with_innertube_base_url(server.base_url())
            .unwrap();
        let output = NamedTempFile::new().unwrap();

        let downloaded = client
            .download_video("https://youtu.be/dQw4w9WgXcQ", output.path())
            .await
            .unwrap();

        assert_eq!(downloaded.format.itag, 22);
        assert_eq!(downloaded.bytes_written, 10);
        assert_eq!(tokio::fs::read(output.path()).await.unwrap(), b"high-video");
    }

    fn test_client() -> YoutubeClient {
        YoutubeClient::with_http_client(reqwest::Client::builder().no_proxy().build().unwrap())
            .unwrap()
    }

    struct TestServer {
        addr: SocketAddr,
    }

    impl TestServer {
        async fn start(body: String) -> Self {
            let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
            let addr = listener.local_addr().unwrap();

            tokio::spawn(async move {
                let (mut socket, _) = listener.accept().await.unwrap();
                let mut buffer = [0; 4096];
                let _ = socket.read(&mut buffer).await.unwrap();
                write_response(&mut socket, "application/json", body.as_bytes()).await;
            });

            Self { addr }
        }

        fn base_url(&self) -> String {
            format!("http://{}/", self.addr)
        }
    }

    #[derive(Clone)]
    struct DownloadServer {
        addr: SocketAddr,
        player_body: Arc<tokio::sync::Mutex<String>>,
    }

    impl DownloadServer {
        async fn start() -> Self {
            let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
            let addr = listener.local_addr().unwrap();
            let player_body = Arc::new(tokio::sync::Mutex::new(String::new()));
            let body_ref = Arc::clone(&player_body);

            tokio::spawn(async move {
                loop {
                    let (mut socket, _) = listener.accept().await.unwrap();
                    let body_ref = Arc::clone(&body_ref);
                    tokio::spawn(async move {
                        let mut buffer = [0; 4096];
                        let bytes_read = socket.read(&mut buffer).await.unwrap();
                        let request = String::from_utf8_lossy(&buffer[..bytes_read]);
                        if request.starts_with("GET /video-high") {
                            write_response(&mut socket, "video/mp4", b"high-video").await;
                        } else if request.starts_with("GET /video-low") {
                            write_response(&mut socket, "video/mp4", b"low-video").await;
                        } else {
                            let body = body_ref.lock().await.clone();
                            write_response(&mut socket, "application/json", body.as_bytes()).await;
                        }
                    });
                }
            });

            Self { addr, player_body }
        }

        async fn set_player_body(&self, body: String) {
            *self.player_body.lock().await = body;
        }

        fn base_url(&self) -> String {
            format!("http://{}/", self.addr)
        }
    }

    async fn write_response(socket: &mut tokio::net::TcpStream, content_type: &str, body: &[u8]) {
        let header = format!(
            "HTTP/1.1 200 OK\r\ncontent-type: {content_type}\r\ncontent-length: {}\r\nconnection: close\r\n\r\n",
            body.len()
        );
        socket.write_all(header.as_bytes()).await.unwrap();
        socket.write_all(body).await.unwrap();
    }
}
