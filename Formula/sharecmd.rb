class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.71/sharecmd_0.0.71_Darwin_x86_64.tar.gz"
  version "0.0.71"
  sha256 "f7074b0825d512bdc625570d23075fda3e759464a08dc4ddaba5868dbd9d6190"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
