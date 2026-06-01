use std::{env, path::PathBuf, process};

use youtube_downloader::YoutubeClient;

#[tokio::main]
async fn main() {
    if let Err(error) = run().await {
        eprintln!("error: {error}");
        process::exit(1);
    }
}

async fn run() -> youtube_downloader::Result<()> {
    let mut args = env::args().skip(1);
    let Some(video_id_or_url) = args.next() else {
        eprintln!("usage: youtube-download <video-id-or-url> [output-file]");
        process::exit(2);
    };
    let output = args
        .next()
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from(format!("{video_id_or_url}.mp4")));

    let client = YoutubeClient::new()?;
    let downloaded = client.download_video(&video_id_or_url, &output).await?;

    println!(
        "downloaded {} bytes to {} using itag {} ({})",
        downloaded.bytes_written,
        output.display(),
        downloaded.format.itag,
        downloaded
            .format
            .quality_label
            .as_deref()
            .unwrap_or("unknown quality")
    );

    Ok(())
}
