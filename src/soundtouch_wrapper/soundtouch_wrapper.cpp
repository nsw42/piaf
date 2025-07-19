// soundtouch_wrapper.cpp
#include <SoundTouch.h>
#include "soundtouch_wrapper.h"

using namespace soundtouch;

void* soundtouch_new(int sampleRate, int channels, float tempo) {
    SoundTouch* st = new SoundTouch();
    st->setSampleRate(sampleRate);
    st->setChannels(channels);
    st->setTempo(tempo);
    st->setSetting(SETTING_AA_FILTER_LENGTH, 8);
    return st;
}

void soundtouch_delete(void* st) {
    delete static_cast<SoundTouch*>(st);
}

void soundtouch_set_tempo(void* st, float tempo) {
    static_cast<SoundTouch*>(st)->setTempo(tempo);
}

void soundtouch_put_samples(void* st, const float* samples, int numSamples) {
    static_cast<SoundTouch*>(st)->putSamples(samples, numSamples);
}

int soundtouch_num_unprocessed_samples(void* st) {
    return static_cast<SoundTouch*>(st)->numUnprocessedSamples();
}

int soundtouch_receive_samples(void* st, float* out, int maxSamples) {
    return static_cast<SoundTouch*>(st)->receiveSamples(out, maxSamples);
}

void soundtouch_flush(void* st) {
    static_cast<SoundTouch*>(st)->flush();
}
