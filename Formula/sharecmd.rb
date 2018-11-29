class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.35/sharecmd_0.0.35_Darwin_x86_64.tar.gz"
  version "0.0.35"
  sha256 "28e2eb3d6c7f0f83e5bcf94f172032fe0a6ae2abda9307daaebaf787ff934fa8"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
