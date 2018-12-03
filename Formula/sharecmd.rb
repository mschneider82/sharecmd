class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.60/sharecmd_0.0.60_Darwin_x86_64.tar.gz"
  version "0.0.60"
  sha256 "f672c527d04c5cfe3059ad3b8cebdb1cbe0ab3e55a6d9c7cd258289d708661c9"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
