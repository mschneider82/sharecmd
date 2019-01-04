class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.86/sharecmd_0.0.86_Darwin_x86_64.tar.gz"
  version "0.0.86"
  sha256 "734968f21f2ff67e903ec291822eb6db17b38d4fffcf027b875580b1c199d5d9"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
