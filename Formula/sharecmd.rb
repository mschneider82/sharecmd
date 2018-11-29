class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.40/sharecmd_0.0.40_Darwin_x86_64.tar.gz"
  version "0.0.40"
  sha256 "7f3b52e0eb33b0d69a984f15553fb8dd302e0ba7c5a7f2d7f03d47c3d40efec7"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
