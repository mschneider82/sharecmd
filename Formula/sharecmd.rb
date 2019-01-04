class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.80/sharecmd_0.0.80_Darwin_x86_64.tar.gz"
  version "0.0.80"
  sha256 "a1caa6644915174a69383ff8172f4af641bd4378461be215300d469c1a0a9b68"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
