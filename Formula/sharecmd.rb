class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.75/sharecmd_0.0.75_Darwin_x86_64.tar.gz"
  version "0.0.75"
  sha256 "69f39d3a4e63114b85d93881967ad1e17061a01eebf0af8a238b9835121062a7"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
