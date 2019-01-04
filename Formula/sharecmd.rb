class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.83/sharecmd_0.0.83_Darwin_x86_64.tar.gz"
  version "0.0.83"
  sha256 "54a815fc07200ebe43fc5b68b7aaee87faf467c5845dc3bc7522a753ff21b14e"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
