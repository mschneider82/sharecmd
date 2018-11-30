class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.46/sharecmd_0.0.46_Darwin_x86_64.tar.gz"
  version "0.0.46"
  sha256 "031a93d080bf9c86f5ecf5711fdeedd0845f8fa6bfe86214f146dc8dbef68c5e"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
