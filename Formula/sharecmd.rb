class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.33/sharecmd_0.0.33_Darwin_x86_64.tar.gz"
  version "0.0.33"
  sha256 "be7ad95acc519bbb715c16b020aab845fe5f46503fbfc9a25fa426cdadea0aea"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
