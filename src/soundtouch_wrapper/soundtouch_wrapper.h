// soundtouch_wrapper.h
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

void* soundtouch_new(int sampleRate, int channels, float tempo);
void soundtouch_delete(void* st);
void soundtouch_set_tempo(void* st, float tempo);
void soundtouch_put_samples(void* st, const float* samples, int numSamples);
int soundtouch_receive_samples(void* st, float* out, int maxSamples);
int soundtouch_num_unprocessed_samples(void* st);
void soundtouch_flush(void* st);

#ifdef __cplusplus
}
#endif
