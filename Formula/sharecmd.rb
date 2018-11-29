class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.28/sharecmd_0.0.28_Darwin_x86_64.tar.gz"
  version "0.0.28"
  sha256 "6b543c1b783d16410ace43254e5f9ebc147425ca1983642e7720784301c3e63b"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
