package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

var rtpPacketCache sync.Map

const (
	rtpPort  = 15006
	rtcpPort = 15007
)

func main() {
	var wg sync.WaitGroup
	wg.Add(3)

	go runRTPSender(&wg)
	go runRTCPSender(&wg)
	go runNackReceiver(&wg)

	wg.Wait()
}

func runRTPSender(wg *sync.WaitGroup) {
	defer wg.Done()

	rtcpListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtcpPort})
	if err != nil {
		fmt.Printf("RTP Sender: failed to listen on RTCP port, %v\n", err)
		return
	}
	defer rtcpListener.Close()

	feedbackChan := make(chan float64, 1)

	go func() {
		buf := make([]byte, 1500)
		for {
			n, _, err := rtcpListener.ReadFromUDP(buf)
			if err != nil {
				fmt.Printf("RTP Sender: failed to read from UDP: %v\n", err)
				return
			}

			packets, err := rtcp.Unmarshal(buf[:n])
			if err != nil {
				fmt.Printf("RTP Sender: failed to unmarshal RTCP packets: %v\n", err)
				continue
			}

			for _, p := range packets {
				if nack, ok := p.(*rtcp.TransportLayerNack); ok {
					fmt.Printf("RTP Sender: received NACK from SSRC %d\n", nack.SenderSSRC)

					for _, pair := range nack.Nacks {
						pid := pair.PacketID
						if val, ok := rtpPacketCache.Load(pid); ok {
							packet := val.(*rtp.Packet)
							buf, err := packet.Marshal()
							if err != nil {
								fmt.Printf("RTP Sender: failed to marshal retransmitted packet: %v\n", err)
								continue
							}
							remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtpPort}
							conn, err := net.DialUDP("udp", nil, remoteAddr)
							if err != nil {
								fmt.Printf("RTP Sender: failed to dial UDP for retransmission, %v\n", err)
								continue
							}
							_, err = conn.Write(buf)
							conn.Close()
							if err != nil {
								fmt.Printf("RTP Sender: failed to write retransmitted packet to UDP: %v\n", err)
							} else {
								fmt.Printf("RTP Sender: retransmitted packet with SequenceNumber %d\n", packet.SequenceNumber)
							}
						}
					}
				} else if rr, ok := p.(*rtcp.ReceiverReport); ok {
					if len(rr.Reports) > 0 {
						report := rr.Reports[0]
						fractionLost := float64(report.FractionLost) / 256.0
						fmt.Printf("RTP Sender: received Receiver Report. Fraction Lost: %.2f%%\n", fractionLost*100)
						feedbackChan <- fractionLost
					}
				}
			}
		}
	}()

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtpPort}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Printf("RTP Sender: failed to dial UDP, %v\n", err)
		return
	}
	defer conn.Close()

	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    0,
			SSRC:           147258,
			Timestamp:      0,
			SequenceNumber: 0,
		},
		Payload: []byte{1, 2, 3, 4, 5, 6, 7, 8}, // 增加 payload 长度以更好地模拟
	}

	const fecGroupSize = 4
	var fecGroup []*rtp.Packet
	fecPayloadType := uint8(100)

	currentBitrateKbps := 100.0
	interval := time.Duration(1000/currentBitrateKbps) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case fractionLost := <-feedbackChan:
			if fractionLost > 0.05 {
				currentBitrateKbps *= 0.8
				fmt.Printf("RTP Sender: high packet loss detected, lowering bitrate to %.2f kbps\n", currentBitrateKbps)
			} else if fractionLost < 0.01 {
				currentBitrateKbps *= 1.1
				fmt.Printf("RTP Sender: low packet loss detected, increasing bitrate to %.2f kbps\n", currentBitrateKbps)
			}
			if currentBitrateKbps < 50 {
				currentBitrateKbps = 50
			}
			if currentBitrateKbps > 500 {
				currentBitrateKbps = 500
			}
			newInterval := time.Duration(1000/currentBitrateKbps) * time.Millisecond
			ticker.Reset(newInterval)

		case <-ticker.C:
			packet.SequenceNumber++
			packet.Timestamp += 160

			packetCopy := &rtp.Packet{}
			*packetCopy = *packet
			rtpPacketCache.Store(packetCopy.SequenceNumber, packetCopy)
			fecGroup = append(fecGroup, packetCopy)

			buf, err := packet.Marshal()
			if err != nil {
				fmt.Printf("RTP Sender: failed to marshal packet: %v\n", err)
				continue
			}
			_, err = conn.Write(buf)
			if err != nil {
				fmt.Printf("RTP Sender: failed to write to UDP: %v\n", err)
				return
			}
			fmt.Printf("RTP Sender: sent data packet with SequenceNumber %d\n", packet.SequenceNumber)

			if len(fecGroup) == fecGroupSize {
				fecPayload := make([]byte, len(fecGroup[0].Payload))
				for _, p := range fecGroup {
					for i := 0; i < len(p.Payload); i++ {
						fecPayload[i] ^= p.Payload[i]
					}
				}

				fecPacket := &rtp.Packet{
					Header: rtp.Header{
						Version:        2,
						PayloadType:    fecPayloadType,
						SSRC:           packet.SSRC,
						Timestamp:      packet.Timestamp,
						SequenceNumber: packet.SequenceNumber + 1,
					},
					Payload: fecPayload,
				}
				rtpPacketCache.Store(fecPacket.SequenceNumber, fecPacket)

				fecBuf, err := fecPacket.Marshal()
				if err != nil {
					fmt.Printf("RTP Sender: failed to marshal FEC packet: %v\n", err)
					continue
				}

				_, err = conn.Write(fecBuf)
				if err != nil {
					fmt.Printf("RTP Sender: failed to write FEC packet to UDP: %v\n", err)
				} else {
					fmt.Printf("RTP Sender: sent FEC packet with SequenceNumber %d\n", fecPacket.SequenceNumber)
				}
				fecGroup = nil
			}
		}
	}
}

func runRTCPSender(wg *sync.WaitGroup) {
	defer wg.Done()

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtcpPort}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Printf("RTCP Sender: failed to dial UDP, %v\n", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ssrc := uint32(147259)
	ntpEpoch := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	rtpTime := uint32(0)
	packetCount := uint32(0)
	octetCount := uint32(0)

	for range ticker.C {
		now := time.Now().UTC()
		ntpTime := uint64(now.Sub(ntpEpoch).Seconds() * (1 << 32))

		sr := &rtcp.SenderReport{
			SSRC:        ssrc,
			NTPTime:     ntpTime,
			RTPTime:     rtpTime,
			PacketCount: packetCount,
			OctetCount:  octetCount,
		}

		packetCount += 25
		octetCount += 25 * 4
		rtpTime += 160 * 25

		buf, err := sr.Marshal()
		if err != nil {
			fmt.Printf("RTCP Sender: failed to marshal SR: %v\n", err)
			continue
		}

		_, err = conn.Write(buf)
		if err != nil {
			fmt.Printf("RTCP Sender: failed to write to UDP: %v\n", err)
			return
		}
		fmt.Printf("RTCP Sender: sent Sender Report from SSRC %d\n", sr.SSRC)
	}
}

func runNackReceiver(wg *sync.WaitGroup) {
	defer wg.Done()

	rtpListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtpPort})
	if err != nil {
		fmt.Printf("NACK Receiver: failed to listen on RTP port, %v\n", err)
		return
	}
	defer rtpListener.Close()

	const fecGroupSize = 4
	fecPayloadType := uint8(100)
	rtpCache := make(map[uint16]*rtp.Packet)

	var expectedSN uint16 = 0
	buf := make([]byte, 1500)

	nackRemoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: rtcpPort}
	nackConn, err := net.DialUDP("udp", nil, nackRemoteAddr)
	if err != nil {
		fmt.Printf("NACK Receiver: failed to dial UDP for NACK, %v\n", err)
		return
	}
	defer nackConn.Close()

	var receivedPackets uint32 = 0
	var lostPackets uint32 = 0
	var lastReportSent time.Time = time.Now()
	const reportInterval = 5 * time.Second

	for {
		n, _, err := rtpListener.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("NACK Receiver: failed to read from UDP: %v\n", err)
			return
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			fmt.Printf("NACK Receiver: failed to unmarshal RTP packet: %v\n", err)
			continue
		}

		// 模拟丢包
		if packet.SequenceNumber%20 == 0 {
			fmt.Printf("NACK Receiver: simulating loss of packet with SequenceNumber %d\n", packet.SequenceNumber)
			lostPackets++
			continue
		}

		rtpCache[packet.SequenceNumber] = packet

		if packet.PayloadType == fecPayloadType {
			fmt.Printf("NACK Receiver: received FEC packet with SequenceNumber %d\n", packet.SequenceNumber)

			if len(rtpCache) >= fecGroupSize {
				lostSN := uint16(0)
				for i := uint16(0); i < fecGroupSize; i++ {
					if _, ok := rtpCache[packet.SequenceNumber-1-i]; !ok {
						lostSN = packet.SequenceNumber - 1 - i
						break
					}
				}

				if lostSN != 0 {
					fmt.Printf("NACK Receiver: attempting to recover lost packet %d using FEC\n", lostSN)

					recoveredPayload := make([]byte, len(packet.Payload))
					copy(recoveredPayload, packet.Payload)

					for i := uint16(0); i < fecGroupSize; i++ {
						currentSN := packet.SequenceNumber - 1 - i
						if currentSN != lostSN {
							if p, ok := rtpCache[currentSN]; ok {
								for j := 0; j < len(p.Payload); j++ {
									recoveredPayload[j] ^= p.Payload[j]
								}
							}
						}
					}
					recoveredPacket := &rtp.Packet{
						Header: rtp.Header{
							PayloadType:    0,
							SSRC:           packet.SSRC,
							SequenceNumber: lostSN,
						},
						Payload: recoveredPayload,
					}
					rtpCache[lostSN] = recoveredPacket
					fmt.Printf("NACK Receiver: successfully recovered packet with SequenceNumber %d\n", lostSN)
					lostPackets--
				}
			}
		} else {
			fmt.Printf("NACK Receiver: received data packet with SequenceNumber %d\n", packet.SequenceNumber)
			receivedPackets++

			if packet.SequenceNumber > expectedSN+1 {
				fmt.Printf("NACK Receiver: detected loss! Expected %d, but got %d\n", expectedSN+1, packet.SequenceNumber)

				nackPacketSN := expectedSN + 1
				// 检查FEC包是否已收到，如果已收到则不发送NACK
				if _, ok := rtpCache[nackPacketSN+4]; ok {
					fmt.Printf("NACK Receiver: FEC packet for %d received, waiting for recovery\n", nackPacketSN)
				} else {
					nack := &rtcp.TransportLayerNack{
						SenderSSRC: packet.SSRC,
						MediaSSRC:  packet.SSRC,
						Nacks: []rtcp.NackPair{
							{
								PacketID:    expectedSN + 1,
								LostPackets: 0,
							},
						},
					}

					nackBuf, err := nack.Marshal()
					if err != nil {
						fmt.Printf("NACK Receiver: failed to marshal NACK packet: %v\n", err)
						continue
					}

					_, err = nackConn.Write(nackBuf)
					if err != nil {
						fmt.Printf("NACK Receiver: failed to write NACK packet to UDP: %v\n", err)
					} else {
						fmt.Printf("NACK Receiver: sent NACK for packet with SequenceNumber %d\n", expectedSN+1)
					}
				}
			}
			expectedSN = packet.SequenceNumber
		}

		if time.Since(lastReportSent) >= reportInterval {
			var fractionLost uint8 = 0
			if receivedPackets > 0 {
				fractionLost = uint8(float64(lostPackets) / float64(receivedPackets+lostPackets) * 256)
			}

			rr := &rtcp.ReceiverReport{
				SSRC: 147259,
				Reports: []rtcp.ReceptionReport{
					{
						SSRC:               packet.SSRC,
						FractionLost:       fractionLost,
						LastSequenceNumber: uint32(packet.SequenceNumber),
					},
				},
			}

			rrBuf, err := rr.Marshal()
			if err != nil {
				fmt.Printf("NACK Receiver: failed to marshal RR packet: %v\n", err)
				continue
			}

			_, err = nackConn.Write(rrBuf)
			if err != nil {
				fmt.Printf("NACK Receiver: failed to write RR packet to UDP: %v\n", err)
			} else {
				fmt.Printf("NACK Receiver: sent Receiver Report with Fraction Lost: %.2f%%\n", float64(fractionLost)/256.0*100)
			}
			receivedPackets = 0
			lostPackets = 0
			lastReportSent = time.Now()
		}
	}
}
